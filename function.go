package function

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"arnested.dk/go/dsupdate"
)

// Handle is the entrypoint for the Google Cloud Function.
func Handle(w http.ResponseWriter, r *http.Request) {
	// We log to Google Cloud Functions and don't need a timestamp
	// since it will be present in the log anyway. On the other
	// hand a reference to file and line number would be nice.
	log.SetFlags(log.Lshortfile)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if !isAuthorized(r.URL.Query().Get("token")) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	payload, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	eventName, err := dnsimpleEventName(payload)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if eventName != "dnssec.rotation_start" && eventName != "dnssec.rotation_complete" {
		http.Error(w, "Not a rotation event", http.StatusOK)

		return
	}

	event, err := dnsimpleEvent(payload)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	config, err := getConfig(event.DelegationSignerRecord.DomainID)

	if err != nil {
		log.Printf("No DK Hostmaster config for %d: %s", event.DelegationSignerRecord.DomainID, err.Error())
		// It's OK if there is no configuration. It could be a
		// domain not handled by DK Hostmaster and/or DNSSEC.
		// We send a 200 OK so DNSimple will not retry.
		http.Error(w, "Missing DK Hostmaster credentials config", http.StatusOK)

		return
	}

	client, err := dsupdate.New(dsupdate.Credentials{
		Domain:   config.Domain,
		UserID:   config.UserID,
		Password: config.Password,
	})

	if err != nil {
		log.Printf("Could not create DSUpdate object: %s", err.Error())
		http.Error(w, "Could not create DSUpdate object", http.StatusInternalServerError)

		return
	}

	dnsimpleToken, ok := os.LookupEnv("DNSIMPLE_TOKEN")

	if !ok {
		http.Error(w, "Missing DNSimple token", http.StatusUnprocessableEntity)

		return
	}

	records, err := dsRecords(dnsimpleToken, config.Domain)

	if err != nil {
		log.Printf("Could not get DS records from DNSimple: %s", err.Error())
		http.Error(w, "Could not get DS records from DNSimple", http.StatusInternalServerError)

		return
	}

	for _, record := range records {
		keyTag, _ := strconv.ParseUint(record.Keytag, 10, 16)
		algorithm, _ := strconv.ParseUint(record.Algorithm, 10, 8)
		digestType, _ := strconv.ParseUint(record.DigestType, 10, 8)

		client.Add(dsupdate.DsRecord{
			KeyTag:     uint16(keyTag),
			Algorithm:  uint8(algorithm),
			DigestType: uint8(digestType),
			Digest:     record.Digest,
		})
	}

	resp, err := client.Post(http.Client{})

	if err != nil {
		log.Printf("Could not update DS records: %s", err.Error())
		http.Error(w, "Could not update DS records", http.StatusInternalServerError)

		return
	}

	_, _ = w.Write(resp)
}

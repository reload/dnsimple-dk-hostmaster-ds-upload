package function

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dnsimple/dnsimple-go/dnsimple/webhook"
)

// Handle is the entrypoint for the Google Cloud Function.
func Handle(w http.ResponseWriter, r *http.Request) {
	// We log to Google Cloud Functions and don't need a timestamp
	// since it will be present in the log anyway. On the other
	// hand a reference to file and line number would be nice.
	log.SetFlags(log.Lshortfile)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if !isAuthorized(r.URL.Query().Get("token")) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	defer r.Body.Close()
	payload, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Printf("Could not read HTTP Body: %s", err.Error())
		http.Error(w, "Could not read HTTP Body", http.StatusBadRequest)

		return
	}

	name, err := webhook.ParseName(payload)

	if err != nil {
		log.Printf("Could not parse event name from payload: %s", err.Error())
		http.Error(w, "Could not parse the event name", http.StatusBadRequest)

		return
	}

	if name != "dnssec.rotation_complete" {
		log.Printf("Not a `dnssec.rotation_complete` event: %s", name)
		// It's OK that this is not the event we are looking
		// for. We send a 200 OK so DNSimple will not retry.
		http.Error(w, "Not a `dnssec.rotation_complete` event", http.StatusOK)

		return
	}

	event := &webhook.DNSSECEvent{}
	err = webhook.ParseDNSSECEvent(event, payload)

	if err != nil {
		log.Printf("Could not parse event as a DNSSEC event: %s", err.Error())
		http.Error(w, "Could not parse event as a DNSSEC event", http.StatusBadRequest)

		return
	}

	// We get the config for DK Hostmaster from the
	// environment. They are prefixed with DNSimple domain ID.
	config, err := getConfig(event.DelegationSignerRecord.DomainID)

	if err != nil {
		log.Printf("No DK Hostmaster config for %d: %s", event.DelegationSignerRecord.DomainID, err.Error())
		// It's OK if there is no configuration. It could be a
		// domain not handled by DK Hostmaster.and/or DNSSEC.
		// We send a 200 OK so DNSimple will not retry.
		http.Error(w, "Missing DK Hostmaster credentials", http.StatusOK)

		return
	}

	body, err := dsUpload(config, event.DelegationSignerRecord)

	if err != nil {
		log.Printf("Upload problem: %s", err.Error())
		http.Error(w, "Internal server error uploading to DK Hostmaster", http.StatusInternalServerError)

		return
	}

	log.Printf("Upload succeeded: %s", body)
	_, _ = w.Write(body)
}

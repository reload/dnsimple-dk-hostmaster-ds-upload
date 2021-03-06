package function

import (
	"context"
	"errors"
	"strconv"

	"arnested.dk/go/dsupdate"
	"github.com/dnsimple/dnsimple-go/dnsimple"
	"golang.org/x/oauth2"
)

func dsRecords(oauthToken string, domain string) ([]dsupdate.DsRecord, error) {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: oauthToken})
	tc := oauth2.NewClient(ctx, ts)

	// new client
	client := dnsimple.NewClient(tc)

	// get the current authenticated account (if you don't know who you are)
	whoamiResponse, err := client.Identity.Whoami(ctx)
	if err != nil {
		return nil, err
	}

	// either assign the account ID or fetch it from the response
	// if you are authenticated with an account token
	accountID := strconv.FormatInt(whoamiResponse.Data.Account.ID, 10)

	// get the list of domains
	// domainsResponse, err := client.Domains.ListDomains(accountID, nil)
	//	domainsResponse, err := client.Domains.ListDomains(accountID, &dnsimple.DomainListOptions{NameLike: domain})
	dsRecords, err := client.Domains.ListDelegationSignerRecords(ctx, accountID, domain, &dnsimple.ListOptions{})
	if err != nil {
		return nil, err
	}

	if len(dsRecords.Data) < 1 {
		return nil, errors.New("No DS records found")
	}

	records := []dsupdate.DsRecord{}

	for _, record := range dsRecords.Data {
		keyTag, _ := strconv.ParseUint(record.Keytag, 10, 16)
		algorithm, _ := strconv.ParseUint(record.Algorithm, 10, 8)
		digestType, _ := strconv.ParseUint(record.DigestType, 10, 8)

		records = append(records, dsupdate.DsRecord{
			KeyTag:     uint16(keyTag),
			Algorithm:  uint8(algorithm),
			DigestType: uint8(digestType),
			Digest:     record.Digest,
		})
	}

	return records, nil
}

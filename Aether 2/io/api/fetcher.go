// API > Fetcher
// This file implements the methods that fetch the data from remotes. Mind that this is only for fetching, the lifecycle and the checks on whether the remote node is available for fetching is handled in dispatch. It deals with getting the data in, it does not deal with decisions on what actions to take (intro, update, search), neither it does with what method to use (get, post).

package api

import (
	"aether-core/services/fingerprinting"
	"aether-core/services/globals"
	"aether-core/services/logging"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Exists checks whether a given item exists in the current DB. This is here because we cannot import persistence due to import cycle being formed, and this is the only place this is being used.
func ExistsInDB(entityType string, fp Fingerprint, lu Timestamp) bool {
	var tableName string
	var result bool
	if entityType == "board" {
		tableName = "Boards"
	} else if entityType == "thread" {
		tableName = "Threads"
	} else if entityType == "post" {
		tableName = "Posts"
	} else if entityType == "vote" {
		tableName = "Votes"
	} else if entityType == "key" {
		tableName = "PublicKeys"
	} else if entityType == "truststate" {
		tableName = "Truststates"
	} else {
		logging.Log(1, fmt.Sprintf("ExistsInDB does not support the entity type you provided. You provided: %s", entityType))
		return false
	}
	query, args, err := sqlx.In(fmt.Sprintf("SELECT count(1) FROM %s WHERE Fingerprint IN (?) AND LastUpdate >= (?);", tableName), fp, lu)
	if err != nil {
		logging.Log(1, fmt.Sprintf("ExistsInDB errored out. Error: %s", err))
		return false
	}
	rows, err := globals.DbInstance.Queryx(query, args...)
	defer rows.Close() // In case of premature exit.
	if err != nil {
		logging.Log(1, fmt.Sprintf("ExistsInDB errored out. Error: %s\n", err))
		return false
	}
	for rows.Next() {
		err = rows.Scan(&result)
		if err != nil {
			logging.Log(1, fmt.Sprintf("ExistsInDB errored out. Error: %s\n", err))
			return false
		}
	}
	rows.Close()
	return result
}

func InsertApiResponseToResponse(response Response, apiresp ApiResponse) Response {
	response.AddressIndexes = apiresp.ResponseBody.AddressIndexes
	response.Addresses = apiresp.ResponseBody.Addresses
	response.BoardIndexes = apiresp.ResponseBody.BoardIndexes
	response.Boards = apiresp.ResponseBody.Boards
	response.KeyIndexes = apiresp.ResponseBody.KeyIndexes
	response.Keys = apiresp.ResponseBody.Keys
	response.PostIndexes = apiresp.ResponseBody.PostIndexes
	response.Posts = apiresp.ResponseBody.Posts
	response.ThreadIndexes = apiresp.ResponseBody.ThreadIndexes
	response.Threads = apiresp.ResponseBody.Threads
	response.TruststateIndexes = apiresp.ResponseBody.TruststateIndexes
	response.Truststates = apiresp.ResponseBody.Truststates
	response.VoteIndexes = apiresp.ResponseBody.VoteIndexes
	response.Votes = apiresp.ResponseBody.Votes
	response.CacheLinks = apiresp.Results
	if response.MostRecentSourceTimestamp < apiresp.Timestamp {
		response.MostRecentSourceTimestamp = apiresp.Timestamp
	}
	return response
}

func concatResponses(response Response, response2 Response) Response {
	response.AddressIndexes = append(
		response.AddressIndexes, response2.AddressIndexes...)
	response.Addresses = append(
		response.Addresses, response2.Addresses...)
	response.BoardIndexes = append(
		response.BoardIndexes, response2.BoardIndexes...)
	response.Boards = append(
		response.Boards, response2.Boards...)
	response.KeyIndexes = append(
		response.KeyIndexes, response2.KeyIndexes...)
	response.Keys = append(
		response.Keys, response2.Keys...)
	response.PostIndexes = append(
		response.PostIndexes, response2.PostIndexes...)
	response.Posts = append(
		response.Posts, response2.Posts...)
	response.ThreadIndexes = append(
		response.ThreadIndexes, response2.ThreadIndexes...)
	response.Threads = append(
		response.Threads, response2.Threads...)
	response.TruststateIndexes = append(
		response.TruststateIndexes, response2.TruststateIndexes...)
	response.Truststates = append(
		response.Truststates, response2.Truststates...)
	response.VoteIndexes = append(
		response.VoteIndexes, response2.VoteIndexes...)
	response.Votes = append(
		response.Votes, response2.Votes...)
	if response.MostRecentSourceTimestamp < response2.MostRecentSourceTimestamp {
		response.MostRecentSourceTimestamp = response2.MostRecentSourceTimestamp
	}
	return response
}

// Basic, reusable instances of transport and client.

// var transport = &http.Transport{
// // TODO: TLS configuration for HTTPS.
// }
var d net.Dialer
var t http.Transport
var c http.Client

// Fetch is the most basic access method. It returns bytes. This should almost never be called directly outside this package.
func Fetch(host string, subhost string, port uint16, location string, method string, postBody []byte) ([]byte, error) {
	// Gotcha of setting these here, these will be repeated every time this is called. Maybe we can run this somehow one time...
	dialer := &d
	dialer.Timeout = globals.BackendConfig.GetTCPConnectTimeout()
	// Dialer configuration inserted here.
	t.Dial = dialer.Dial
	t.TLSHandshakeTimeout = globals.BackendConfig.GetTLSHandshakeTimeout()
	transport := &t
	// Transport configuration settings inserted here.
	c.Transport = transport
	c.Timeout = globals.BackendConfig.GetConnectionTimeout()
	client := &c

	// fmt.Println(client.Timeout)
	// fmt.Println(globals.ConnectionTimeout)
	var fullLink string
	if len(subhost) > 0 {
		fullLink = fmt.Sprint(
			"http://", host, ":", strconv.Itoa(int(port)), "/", subhost, "/v0/", location) // TODO: Move to HTTPS after that portion goes live.
	} else {
		fullLink = fmt.Sprint(
			"http://", host, ":", strconv.Itoa(int(port)), "/v0/", location) // TODO: Move to HTTPS after that portion goes live.
	}
	logging.Log(2, fmt.Sprintf("Fetch is being called for the URL: %s", fullLink))
	// TODO: When we have the local profile, the v0 should be coming from the appropriate version number. Constant for the time being.
	var err error
	var resp *http.Response
	if method == "GET" {
		resp, err = client.Get(fullLink)
	} else if method == "POST" {
		resp, err = client.Post(fullLink, "application/json", bytes.NewReader(postBody))
	} else {
		defer resp.Body.Close()
		return []byte{}, errors.New("Unsupported HTTP method. Available methods are: GET, POST")
	}
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return []byte{}, errors.New(
				fmt.Sprint(
					"The host refused the connection. Host:", host,
					", Subhost: ", subhost,
					", Port: ", port,
					", Location: ", location))
		} else if strings.Contains(err.Error(), "Client.Timeout exceeded while awaiting headers") {
			return []byte{}, errors.New(
				fmt.Sprint(
					"Timeout exceeded. Host:", host,
					", Subhost: ", subhost,
					", Port: ", port,
					", Location: ", location))
		} else if strings.Contains(err.Error(), "i/o timeout") {
			return []byte{}, errors.New(
				fmt.Sprint(
					"I/O timeout. Host:", host,
					", Subhost: ", subhost,
					", Port: ", port,
					", Location: ", location))
		} else if strings.Contains(err.Error(), "connection reset by peer") {
			return []byte{}, errors.New(
				fmt.Sprint(
					"Connection reset by peer. Host:", host,
					", Subhost: ", subhost,
					", Port: ", port,
					", Location: ", location))
		} else if strings.Contains(err.Error(), "EOF") {
			return []byte{}, errors.New(
				fmt.Sprint(
					"The remote crashed or shutting down. Host:", host,
					", Subhost: ", subhost,
					", Port: ", port,
					", Location: ", location))
		} else {
			fmt.Println("Fatal error in api.Fetch. Quitting.")
			fmt.Println(err)
			logging.LogCrash(err)
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// logging.LogCrash(err)
			fmt.Sprint(err.Error())
		}
		return body, nil
	} else {
		logging.Log(2, fmt.Sprintf("FULL LINK IN FETCH FOR THIS FAILED REQUEST: \n%s\n", fullLink))
		return []byte{}, errors.New(
			fmt.Sprint(
				"Non-200 status code returned from Fetch. Received status code: ", resp.StatusCode,
				", Host: ", host,
				", Subhost: ", subhost,
				", Port: ", port,
				", Location: ", location,
				", Method: ", method))
	}
	return []byte{}, errors.New("This should never have happened.")
}

// GetPageRaw returns a raw page from the cache. This returns the entire page, not just the data. This is useful for functions that need to be aware of the page's metadata.
func GetPageRaw(host string, subhost string, port uint16, location string, method string, postBody []byte) (ApiResponse, error) {
	// TODO: Kill the connection if the file size is too large, or if it takes too long to download. A page above 5mb is probably malicious, also is one that takes more than 10 minutes to download.
	var apiresp ApiResponse
	result, err := Fetch(host, subhost, port, location, method, postBody)
	if err != nil {
		return apiresp, err
	}
	err2 := json.Unmarshal(result, &apiresp)
	if err2 != nil {
		return apiresp, errors.New(
			fmt.Sprint(
				"The JSON that arrived over the network is malformed. JSON: ", string(result),
				", Host: ", host,
				", Subhost: ", subhost,
				", Port: ", port,
				", Location: ", location))
	}
	// Map over everything you have.
	if method == "POST" {
		logging.Log(2, fmt.Sprintf("We've made a POST request to the endpoint %s and this was its body: %#v", location, string(postBody)))
	}
	pageVerified, err := apiresp.VerifySignature() // If signature check is disabled, this will always return true.
	if err != nil {
		return ApiResponse{}, errors.New(fmt.Sprintf("Page signature verification failed with an error. Error: %s", err))
	}
	if !pageVerified {
		return ApiResponse{}, errors.New("Page signature verification failed. The signature does not match.")
	}
	if len(apiresp.NodePublicKey) > 0 {
		apiresp.NodeId = Fingerprint(fingerprinting.Create(apiresp.NodePublicKey))
	} else {
		/*
			This makes it more obvious that the given field in the database is catchall for all nodes without a node public key. This should never happen in production because by default this check is enabled, and can only be disabled via a command line flag, which forces the app into read-only configs mode.
		*/
		apiresp.NodeId = "NODEID FOR NODE(S) WITH EMPTY NODEPUBLICKEY"
	}
	errs := apiresp.Verify()
	if len(errs) >= 3 {
		errStrs := []string{}
		for _, err := range errs {
			errStrs = append(errStrs, err.Error())
		}
		logging.Log(1, fmt.Sprintf("This page has 3 or more entities who has failed verification. Errors: %#v", errStrs))
		return ApiResponse{}, errors.New(fmt.Sprintf("This page has 3 or more entities who has failed verification"))
	}
	return apiresp, nil
}

// GetPage gets a page from a cache. This returns the data on the provided page.
func GetPage(host string, subhost string, port uint16, location string, method string, postBody []byte) (Response, time.Duration, error) {
	var apiresp ApiResponse
	var response Response
	var start time.Time
	var elapsed time.Duration
	if method == "POST" {
		start = time.Now()
	}
	apiresp, err := GetPageRaw(host, subhost, port, location, method, postBody)
	if err != nil {
		return response, elapsed, err // elapsed is unset, set only on POST below.
	}
	if method == "POST" {
		elapsed = time.Since(start)
	}
	response = InsertApiResponseToResponse(response, apiresp)
	return response, elapsed, nil // elapsed potentially unset.
}

// CollectMultipartPOSTResponse gets the multipart post response from a given post endpoint result. This function does not make the post request, it is only used to collect the result that is generated by a prior POST request.
func CollectMultipartPOSTResponse(host string, subhost string, port uint16, location string) (Response, error) {
	indexResp, err := getIndexOfCache(host, subhost, port, location)
	if err != nil {
		return Response{}, errors.New(fmt.Sprintf("Error raised from GetIndexOfCache inside CollectMultipartPOSTResponse. Error: %s", err))
	}
	// Look at everything in the index and find the things that we want to pull. Page Number : bool pairs help us find which pages to hit.
	allPgs := make(map[int]bool)
	if len(indexResp.BoardIndexes) > 0 {
		for _, val := range indexResp.BoardIndexes {
			if !ExistsInDB("board", val.Fingerprint, val.LastUpdate) {
				// Grab the whole page, DB will remove useless stuff.
				allPgs[val.PageNumber] = true
			}
		}
	}
	if len(indexResp.ThreadIndexes) > 0 {
		for _, val := range indexResp.ThreadIndexes {
			if !ExistsInDB("thread", val.Fingerprint, val.LastUpdate) {
				allPgs[val.PageNumber] = true
			}
		}
	}
	if len(indexResp.PostIndexes) > 0 {
		for _, val := range indexResp.PostIndexes {
			if !ExistsInDB("post", val.Fingerprint, val.LastUpdate) {
				allPgs[val.PageNumber] = true
			}
		}
	}
	if len(indexResp.VoteIndexes) > 0 {
		for _, val := range indexResp.VoteIndexes {
			if !ExistsInDB("vote", val.Fingerprint, val.LastUpdate) {
				allPgs[val.PageNumber] = true
			}
		}
	}
	if len(indexResp.KeyIndexes) > 0 {
		for _, val := range indexResp.KeyIndexes {
			if !ExistsInDB("key", val.Fingerprint, val.LastUpdate) {
				allPgs[val.PageNumber] = true
			}
		}
	}
	if len(indexResp.TruststateIndexes) > 0 {
		for _, val := range indexResp.TruststateIndexes {
			if !ExistsInDB("truststate", val.Fingerprint, val.LastUpdate) {
				allPgs[val.PageNumber] = true
			}
		}
	}
	logging.Log(2, fmt.Sprintf("The pages we have to make a call to are: %#v\n", allPgs))

	// For each page we have for this post response, hit the main cache and gather the data.
	mainResp := Response{}
	for key, _ := range allPgs {
		loc := fmt.Sprint(location, "/", key, ".json")
		logging.Log(2, fmt.Sprintf("Making a request to %s\n", loc))
		resp, _, err := GetPage(host, subhost, port, loc, "GET", []byte{})
		if err != nil {
			return Response{}, err
		}
		mainResp = concatResponses(mainResp, resp)
	}
	return mainResp, nil
}

// GetCache returns an entire cache. This is useful to pull a cache from the remote. This is a single thread process, it does go through the pages in order.  We could bombard the remote with goroutines, but on a larger scale, that would be called a DDoS of the remote node, so we shouldn't do that.
func GetCache(host string, subhost string, port uint16, location string) (Response, error) {
	var response Response
	// Get the first raw page (because we need to access pagination),
	pageResp, err := GetPageRaw(host, subhost, port, fmt.Sprint(location, "/0.json"), "GET", []byte{})
	if err != nil && strings.Contains(err.Error(), "Received status code: 404") {
		return response, errors.New(
			fmt.Sprint(
				"The first page of the cache does not exist. This cache likely does not exist.",
				", Host: ", host,
				", Subhost: ", subhost,
				", Port: ", port,
				", Location: ", location))
	} else if err != nil {
		// If the first page is faulty, bail.
		return response, err
	}
	// And look at the page count, so we know how many times to iterate.
	pageCount := pageResp.Pagination.Pages
	// Convert this raw page response to page response data for merge.
	response = InsertApiResponseToResponse(response, pageResp)
	// Create a counter for missing pages. If 3 of them come one after another, bail.
	brokenPageCounter := 0
	// Iterate over all of the pages, starting from 1 (we already cleared the 0)
	for i := uint64(1); i <= pageCount; i++ { // Pagination starts from 0
		pageResp2, _, err := GetPage(host, subhost, port,
			fmt.Sprint(location, "/", i, ".json"), "GET", []byte{})
		if err != nil {
			logging.Log(1, fmt.Sprintf("GetPage returned this error: Err: %#v", err))
			brokenPageCounter++
			if brokenPageCounter >= 3 {
				logging.Log(1, fmt.Sprint(
					"3 or more broken pages, either missing or verification failures. Stopping the download of this cache.",
					", Host: ", host,
					", Subhost: ", subhost,
					", Port: ", port,
					", Location: ", location,
					", Last page number: ", i))
				return response, errors.New(
					fmt.Sprint(
						"3 or more broken pages, either missing or verification failures. Stopping the download of this cache.",
						", Host: ", host,
						", Subhost: ", subhost,
						", Port: ", port,
						", Location: ", location,
						", Last page number: ", i))
			}
		}
		// And save into the response.
		response = concatResponses(response, pageResp2)
	}
	return response, nil
}

// mapEndpointToEndpointAddress generates the address that needs to be called for the endpoint that is being requested.
func mapEndpointToEndpointAddress(endpoint string) string {
	endpointsMap := map[string]string{
		"boards":      "c0/boards",
		"threads":     "c0/threads",
		"posts":       "c0/posts",
		"votes":       "c0/votes",
		"addresses":   "addresses", // Addresses is a mim entity, not a c0 entity.
		"keys":        "c0/keys",
		"truststates": "c0/truststates"}
	epAddress := endpointsMap[endpoint]
	// If we don't know which endpoint this is, attempt to call it directly.
	if epAddress == "" {
		epAddress = endpoint
	}
	return epAddress
}

// GetEndpoint returns an entire endpoint from the remote node.
func GetEndpoint(host string, subhost string, port uint16, endpoint string, lastCheckin Timestamp) (Response, error) {
	// This is where the mapping for an endpoint to its respective subprotocol folder is mapped. Below this level, you have to supply your own subprotocol string.
	logging.Log(2, fmt.Sprintf("GetEndpoint was called for the endpoint: %s", endpoint))
	epAddress := mapEndpointToEndpointAddress(endpoint)
	var response Response
	// Get raw page, because we need to access index links.
	result, err := getIndexOfEndpoint(host, subhost, port, epAddress)
	// Map the timestamp of the index onto the response we're generating, in case we might not have any caches (this can happen if our internal timestamp for this cache is newer than the last cache's timestamp.)
	response.MostRecentSourceTimestamp = result.MostRecentSourceTimestamp
	indexes := result.CacheLinks
	if err != nil {
		return response, errors.New(
			fmt.Sprint(
				"Get Endpoint failed because it couldn't get the index of the endpoint.",
				", Error: ", err,
				", Host: ", host,
				", Subhost: ", subhost,
				", Port: ", port,
				", Endpoint: ", endpoint))
	}
	// A broken cache can happen because the cache has underlying missing pages, or pages that has failed verification. At the level of the endpoint, it does not matter why the cache has failed, only that it failed. if there are enough failures, we bail.
	brokenCacheCounter := 0
	for _, val := range indexes {
		// If the cache does end after our last checkin timestamp, we want to read that cache.
		// ----------------- Why? -------------------------
		// Example:
		// Assume lastcheckin is 5
		// Assume caches are: 1-2, 2-3, 3-4, 4-5, 5-6, 6-7.
		// We want 4-5, 5-6, 6-7.
		// 5 6 7 (ends)
		// 5,6,7 > lastcheckin = true.
		// ------------------------------------------------
		if val.EndsAt >= lastCheckin {
			// Get the first page of the cache.
			cache, err := GetCache(host, subhost, port,
				fmt.Sprint(epAddress, "/", val.ResponseUrl))
			response = concatResponses(response, cache)
			if err != nil {
				brokenCacheCounter++ // We never reset this within this endpoint call.
				if brokenCacheCounter >= 3 {
					return response, errors.New(
						fmt.Sprint(
							"3 or more cache failures in the same endpoint. Stopping the download of this endpoint.",
							", Error: ", err,
							", Host: ", host,
							", Subhost: ", subhost,
							", Port: ", port,
							", Endpoint: ", endpoint,
							", Cache link: ", fmt.Sprint(endpoint, "/", val.ResponseUrl)))
				}
			}
		}

	}
	boardCount := len(response.Boards)
	threadCount := len(response.Threads)
	postCount := len(response.Posts)
	voteCount := len(response.Votes)
	addressCount := len(response.Addresses)
	keysCount := len(response.Keys)
	truststatesCount := len(response.Truststates)
	// logging.Log(1, fmt.Sprintf("Response for the endpoint %s was %#v\n", endpoint, response))
	logging.Log(2, fmt.Sprintf("GetEndpoint returned for the endpoint: %s. Number of items: Boards: %d, Threads: %d, Posts: %d, Votes: %d, Addresses: %d, Keys: %d, Truststates: %d", endpoint, boardCount, threadCount, postCount, voteCount, addressCount, keysCount, truststatesCount))

	return response, nil
}

// GetRemoteNode downloads the entire remote node data by hitting all endpoints and all caches and all pages within them. This is the bootstrap function. This should be used when the local database is empty and the remote node is new. Never call this when the local database is not empty as that is fairly wasteful.
func GetRemoteNode(host string, subhost string, port uint16) (Response, error) {
	endpoints := []string{
		"boards", "threads", "posts", "votes", "addresses", "keys", "truststates"}
	var response Response
	for _, endpoint := range endpoints {
		resp, err := GetEndpoint(host, subhost, port, endpoint, 0)
		response = concatResponses(response, resp)
		if err != nil {
			// GetRemoteNode continues to work under all conditions. It won't stop the sequence for any errors.
			// NOOP for now.
		}
	}
	return response, nil // It won't communicate out any errors.
}

// getIndexOfCache gets the index of a cache. Location is the url up to cache name.
func getIndexOfCache(
	host string, subhost string, port uint16, location string) (Response, error) {
	firstIndexPage, err := GetPageRaw(
		host, subhost, port, fmt.Sprint(location, "/index/0.json"), "GET", []byte{})
	if err != nil {
		return Response{}, err
	}
	var resp Response
	resp = InsertApiResponseToResponse(resp, firstIndexPage)
	if firstIndexPage.Pagination.Pages > 0 {
		for i := uint64(1); i <= firstIndexPage.Pagination.Pages; i++ {
			page, err := GetPageRaw(host, subhost, port,
				fmt.Sprint(location, "/index/", i, ".json"), "GET", []byte{})
			if err != nil {
				return Response{}, err
			}
			var pgResp Response
			pgResp = InsertApiResponseToResponse(pgResp, page)
			resp = concatResponses(resp, pgResp)
		}
	}
	return resp, nil
}

// getIndexOfEndpoint gets the cache index of an endpoint.
func getIndexOfEndpoint(
	host string, subhost string, port uint16, endpoint string) (Response, error) {
	EndpointIndexResponse, err := GetPageRaw(
		host, subhost, port, fmt.Sprint(endpoint, "/index.json"), "GET", []byte{})
	var resp Response
	resp = InsertApiResponseToResponse(resp, EndpointIndexResponse)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// checkForEntityInAnswer is a private function which returns whether an entity exists in a cache result. If so, it returns the entity. If not, it returns nil.
func checkForEntityInAnswer(a Answer, fp Fingerprint, t string) interface{} {
	switch t {
	case "boards":
		var entities []Board
		entities = append(entities, a.Boards...)
		for _, entity := range entities {
			if entity.Fingerprint == fp {
				return entity
			}
		}
	case "threads":
		var entities []Thread
		entities = append(entities, a.Threads...)
		for _, entity := range entities {
			if entity.Fingerprint == fp {
				return entity
			}
		}
	case "posts":
		var entities []Post
		entities = append(entities, a.Posts...)
		for _, entity := range entities {
			if entity.Fingerprint == fp {
				return entity
			}
		}
	case "votes":
		var entities []Vote
		entities = append(entities, a.Votes...)
		for _, entity := range entities {
			if entity.Fingerprint == fp {
				return entity
			}
		}
	case "addresses":
		// Nothing happens, as addresses aren't queryable
	case "keys":
		var entities []Key
		entities = append(entities, a.Keys...)
		for _, entity := range entities {
			if entity.Fingerprint == fp {
				return entity
			}
		}
	case "truststates":
		var entities []Truststate
		entities = append(entities, a.Truststates...)
		for _, entity := range entities {
			if entity.Fingerprint == fp {
				return entity
			}
		}
	}
	return nil
}

// inTimeRange returns true or false based on whether the value given are within the bounds of the given timestamps.
func inTimeRange(oldest Timestamp, newest Timestamp, val Timestamp) bool {
	if val > oldest && val < newest {
		return true
	} else {
		return false
	}
}

// pullFullEntityFromCache returns the item you have requested by fingerprint from the cache you are pointing at. If no result is found, it will return an empty interface. This could be implemented as a normal GetCache and then search, but that requires the entire cache to be downloaded, whereas this method stops and returns as soon as it can.
func pullFullEntityFromCache(cacheUrl string, cachePage int, fingerprint Fingerprint, t string, host string, subhost string, port uint16) (interface{}, error) {
	if cachePage == 0 {
		// If the cache page is zero, the item we need is either in the first page, or we don't know the cache page, so we need to search.

		// Get the first raw page (because we need to access pagination),
		pageResp, err := GetPageRaw(host, subhost, port, fmt.Sprint(cacheUrl, "/0.json"), "GET", []byte{})
		if err != nil {
			return nil, errors.New(
				fmt.Sprint(
					"The item is not found at the location the index points to.",
					", Error: ", err,
					", CacheUrl: ", cacheUrl,
					", CachePage: ", cachePage,
					", Fingerprint: ", fingerprint))
		}
		// And look at the page count, so we know how many times to iterate.
		pageCount := pageResp.Pagination.Pages
		// Check the Answer type object to see whether we have it or not.
		entity := checkForEntityInAnswer(pageResp.ResponseBody, fingerprint, t)
		if entity == nil {
			// We haven't found what we wanted on the first page, so we go forward on searching other pages.
			for i := uint64(1); i <= pageCount; i++ { // Pagination starts from 0
				pageResp2, err := GetPageRaw(host, subhost, port,
					fmt.Sprint(cacheUrl, "/", i, ".json"), "GET", []byte{})
				if err != nil {
					return nil, errors.New(
						fmt.Sprint(
							"The item is not found at the location the index points to.",
							", Error: ", err,
							", CacheUrl: ", cacheUrl,
							", CachePage: ", cachePage,
							", Fingerprint: ", fingerprint))
				}
				// Again, check for whether entity exists on this page.
				entity := checkForEntityInAnswer(pageResp2.ResponseBody, fingerprint, t)
				if entity != nil {
					// If we have an entity that fits the bill, return it and exit.
					return entity, nil
				}
			}
		} else {
			// If we have found what we want on the first page, return it and exit.
			return entity, nil
		}
	} else {
		// If we know the cache page, we can just directly fetch the item.
		pageResp, err := GetPageRaw(host, subhost, port, fmt.Sprint(cacheUrl, "/", cachePage, ".json"), "GET", []byte{})
		if err != nil {
			return nil, errors.New(
				fmt.Sprint(
					"The item is not found at the location the index points to.",
					", Error: ", err,
					", CacheUrl: ", cacheUrl,
					", CachePage: ", cachePage,
					", Fingerprint: ", fingerprint))
		}
		entity := checkForEntityInAnswer(pageResp.ResponseBody, fingerprint, t)
		return entity, nil
	}

	return nil, errors.New( // If nothing is found, return empty item error.
		fmt.Sprint(
			"The item is not found at the location the index points to.",
			", CacheUrl: ", cacheUrl,
			", CachePage: ", cachePage,
			", Fingerprint: ", fingerprint))
}

// Query struct. Used to provide input to the Query function below.

type QueryData struct {
	EntityType  string
	Fingerprint Fingerprint
	Creation    Timestamp // Can be empty.
	LastUpdate  Timestamp // Last *known* update, can be empty
}

// Query requests an entity from the remote provided. It takes an index form struct. It only returns the requested item or an empty answer.
func Query(host string, subhost string, port uint16, q QueryData) (Response, error) {
	// TODO: Look at the timestamps (update if present, if not, creation, if not, go linear starting from most recent)
	var r Response
	// Before doing anything else, if the type is thread or post, disable LastUpdate. Those items are not updateable.
	updateFieldEnabled := true
	if q.EntityType == "posts" || q.EntityType == "threads" {
		updateFieldEnabled = false
	}
	epAddress := mapEndpointToEndpointAddress(q.EntityType)
	result, err := getIndexOfEndpoint(host, subhost, port, epAddress)
	endpointIndex := result.CacheLinks
	if err != nil {
		return r, nil
	}
	// Do a range search within all caches that include the last update and creation timestamps. This is where we figure out which caches we need to search.
	var cachesSlice []ResultCache
	for _, cache := range endpointIndex {
		if updateFieldEnabled {
			if inTimeRange(cache.StartsFrom, cache.EndsAt, q.Creation) || inTimeRange(cache.StartsFrom, cache.EndsAt, q.LastUpdate) {
				// This adds the endpoints which declare themselves to be in the time range of either Creation, LastUpdate or both. Mind that the creation will be only on one cache, but there may be more than one update.
				cachesSlice = append(cachesSlice, cache)
				// TODO: If there is a LastUpdate, checking Creation is inefficient as the result found in the cache pointed at by Creation will not be used. will not be used. But for purposes of simplicity and to avoid checking for the corner conditions created by having LastUpdate stopping Creation check, I'm leaving it there to be made more efficient in a future refactoring.
			}
		} else {
			// In the case of posts or threads, there is no update field. In that case, a mistakenly provided update field would expand the search into a location that can't possible have it. This section below is here to guard against that waste of resources.
			if inTimeRange(cache.StartsFrom, cache.EndsAt, q.Creation) {
				cachesSlice = append(cachesSlice, cache)
			}
		}
	}
	if updateFieldEnabled {
		if q.Creation == 0 && q.LastUpdate == 0 {
			// If no data is provided as to when the entity could be, we have to go through all of the data to find it.
			cachesSlice = endpointIndex
		}
	} else { // Same as above, guarding against non-updateable entities.
		if q.Creation == 0 {
			cachesSlice = endpointIndex
		}
	} // cachesSlice has all of the caches we have to search now.

CacheIterator: // Naming the for loop CacheIterator.
	for _, cache := range cachesSlice {
		cacheLocation := fmt.Sprint(epAddress, "/", cache.ResponseUrl)
		cIndex, err := getIndexOfCache(host, subhost, port, cacheLocation)
		if err != nil {
			logging.Log(1, fmt.Sprintf("Error in CacheIterator coming from GetIndexOfCache. Error: %s", err))
		}
		// Save the EntityIndexes into proper locations on Response.
		switch q.EntityType {
		case "boards":
			entities := cIndex.BoardIndexes
			// For each of those entities,
			for _, entityIndex := range entities {
				// Check if this is what we want.
				if entityIndex.Fingerprint == q.Fingerprint {
					// If so, pull the result from cache.
					obj, err := pullFullEntityFromCache(cacheLocation, entityIndex.PageNumber, q.Fingerprint, q.EntityType, host, subhost, port)
					if err != nil {
						return r, errors.New(
							fmt.Sprint(
								"Could not pull entity from cache. The item is indexed as available in the remote node, but the actual body of the item is not available.",
								", Error: ", err,
								", Host: ", host,
								", Subhost: ", subhost,
								", Port: ", port,
								", QueryData: ", q))
					}
					// And put into the proper part of the response.
					r.Boards = append(r.Boards, obj.(Board))
					// And finally, break the for loop, so it won't look at other caches when it's done.
					break CacheIterator
				}
			}
		case "threads":
			entities := cIndex.ThreadIndexes
			// For each of those entities,
			for _, entityIndex := range entities {
				// Check if this is what we want.
				if entityIndex.Fingerprint == q.Fingerprint {
					// If so, pull the result from cache.
					obj, err := pullFullEntityFromCache(cacheLocation, entityIndex.PageNumber, q.Fingerprint, q.EntityType, host, subhost, port)
					if err != nil {
						return r, errors.New(
							fmt.Sprint(
								"Could not pull entity from cache. The item is indexed as available in the remote node, but the actual body of the item is not available.",
								", Error: ", err,
								", Host: ", host,
								", Subhost: ", subhost,
								", Port: ", port,
								", QueryData: ", q))
					}
					// And put into the proper part of the response.
					r.Threads = append(r.Threads, obj.(Thread))
					// And finally, break the for loop, so it won't look at other caches when it's done.
					break CacheIterator
				}
			}
		case "posts":
			entities := cIndex.PostIndexes
			// For each of those entities,
			for _, entityIndex := range entities {
				// Check if this is what we want.
				if entityIndex.Fingerprint == q.Fingerprint {
					// If so, pull the result from cache.
					obj, err := pullFullEntityFromCache(cacheLocation, entityIndex.PageNumber, q.Fingerprint, q.EntityType, host, subhost, port)
					if err != nil {
						return r, errors.New(
							fmt.Sprint(
								"Could not pull entity from cache. The item is indexed as available in the remote node, but the actual body of the item is not available.",
								", Error: ", err,
								", Host: ", host,
								", Subhost: ", subhost,
								", Port: ", port,
								", QueryData: ", q))
					}
					// And put into the proper part of the response.
					r.Posts = append(r.Posts, obj.(Post))
					// And finally, break the for loop, so it won't look at other caches when it's done.
					break CacheIterator
				}
			}
		case "votes":
			entities := cIndex.VoteIndexes
			// For each of those entities,
			for _, entityIndex := range entities {
				// Check if this is what we want.
				if entityIndex.Fingerprint == q.Fingerprint {
					// If so, pull the result from cache.
					obj, err := pullFullEntityFromCache(cacheLocation, entityIndex.PageNumber, q.Fingerprint, q.EntityType, host, subhost, port)
					if err != nil {
						return r, errors.New(
							fmt.Sprint(
								"Could not pull entity from cache. The item is indexed as available in the remote node, but the actual body of the item is not available.",
								", Error: ", err,
								", Host: ", host,
								", Subhost: ", subhost,
								", Port: ", port,
								", QueryData: ", q))
					}
					// And put into the proper part of the response.
					r.Votes = append(r.Votes, obj.(Vote))
					// And finally, break the for loop, so it won't look at other caches when it's done.
					break CacheIterator
				}
			}
		case "addresses":
			// Nothing happens, as addresses aren't queryable
			return r, nil
		case "keys":
			entities := cIndex.KeyIndexes
			// For each of those entities,
			for _, entityIndex := range entities {
				// Check if this is what we want.
				if entityIndex.Fingerprint == q.Fingerprint {
					// If so, pull the result from cache.
					obj, err := pullFullEntityFromCache(cacheLocation, entityIndex.PageNumber, q.Fingerprint, q.EntityType, host, subhost, port)
					if err != nil {
						return r, errors.New(
							fmt.Sprint(
								"Could not pull entity from cache. The item is indexed as available in the remote node, but the actual body of the item is not available.",
								", Error: ", err,
								", Host: ", host,
								", Subhost: ", subhost,
								", Port: ", port,
								", QueryData: ", q))
					}
					// And put into the proper part of the response.
					r.Keys = append(r.Keys, obj.(Key))
					// And finally, break the for loop, so it won't look at other caches when it's done.
					break CacheIterator
				}
			}
		case "truststates":
			entities := cIndex.TruststateIndexes
			// For each of those entities,
			for _, entityIndex := range entities {
				// Check if this is what we want.
				if entityIndex.Fingerprint == q.Fingerprint {
					// If so, pull the result from cache.
					obj, err := pullFullEntityFromCache(cacheLocation, entityIndex.PageNumber, q.Fingerprint, q.EntityType, host, subhost, port)
					if err != nil {
						return r, errors.New(
							fmt.Sprint(
								"Could not pull entity from cache. The item is indexed as available in the remote node, but the actual body of the item is not available.",
								", Error: ", err,
								", Host: ", host,
								", Subhost: ", subhost,
								", Port: ", port,
								", QueryData: ", q))
					}
					// And put into the proper part of the response.
					r.Truststates = append(r.Truststates, obj.(Truststate))
					// And finally, break the for loop, so it won't look at other caches when it's done.
					break CacheIterator
				}
			}
		}
	}
	return r, nil
}

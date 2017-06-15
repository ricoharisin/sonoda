package main
 
import (
    "net/http"
    "fmt"
    "io/ioutil"
    "encoding/json"
    "strings"
    "strconv"
)

const mainUrl string = "https://api.github.com/repos/tokopedia/android-tokopedia/pulls"

func main() {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", mainUrl, nil)
	req.Header.Add("Authorization", "token f66d639702c0a0d792251fb7e3c06386d54df629")
	resp, _ := client.Do(req)

	stringResponse := getStringFromResponse(resp)

	checkPullRequest(getByteFromString(stringResponse))

}

func getStringFromResponse(response *http.Response) (string) {
	defer response.Body.Close() 
	body, _ := ioutil.ReadAll(response.Body)

	return string(body)
}

func getByteFromString(origin string) ([]byte) {
	return []byte(origin)
}

func checkPullRequest(dataByte []byte) {
	var data []interface{}
	if err := json.Unmarshal(dataByte, &data); err != nil {
		panic(err)
	}

	for i := range data {
		pullRequest := data[i].(map[string]interface{})

		ableToMerge, number := isPullRequestAbleToMerge(pullRequest)

		if ableToMerge {
			mergePullRequest(number)
		}
	}

}

func getPullRequestReviews(number string) ([]byte) {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", mainUrl+"/"+number+"/reviews", nil)
	req.Header.Add("Authorization", "token f66d639702c0a0d792251fb7e3c06386d54df629")
	resp, _ := client.Do(req)

	stringResponse := getStringFromResponse(resp)

	return getByteFromString(stringResponse)
}

func getPullRequestFilesChanged(number string) ([]byte) {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", mainUrl+"/"+number+"/files", nil)
	req.Header.Add("Authorization", "token f66d639702c0a0d792251fb7e3c06386d54df629")
	resp, _ := client.Do(req)

	stringResponse := getStringFromResponse(resp)

	return getByteFromString(stringResponse)
}

func isFilesChangedValid(number string) (bool) {
	dataByte := getPullRequestFilesChanged(number)

	var data []interface{}
	if err := json.Unmarshal(dataByte, &data); err != nil {
		panic(err)
	}

	if len(data) <= 10 {
		fmt.Println("Pull Request:" +number+ " is files changed valid")
		return true
	} else {
		return false
	}
}

func isHotfix(pullRequest map[string]interface{}, number string) (bool) {
	head := pullRequest["head"].(map[string]interface{})
	base := pullRequest["base"].(map[string]interface{})

	refHead := head["ref"].(string)
	refBase := base["ref"].(string)

	if strings.EqualFold(refBase, "release_candidate") && strings.Contains(strings.ToLower(refHead), "hotfix") {
		fmt.Println("Pull Request:" +number+ " is hotfix")
		return true
	}  else {
		return false
	}
}

func isReviewValid(number string) (bool) {
	dataByte := getPullRequestReviews(number)

	var data []interface{}
	validByQa := false
	numOfReviewer := 0

	if err := json.Unmarshal(dataByte, &data); err != nil {
		panic(err)
	}

	for i := range data {
		pullRequest := data[i].(map[string]interface{})
		state := pullRequest["state"].(string)
		body := pullRequest["body"].(string)
		if strings.EqualFold(state, "APPROVED") {
			if strings.Contains(strings.ToLower(body), "passed qa") {
				validByQa = true
			}
			numOfReviewer++;
		}
	}

	if numOfReviewer >= 2 && validByQa {
		fmt.Println("Pull Request:" +number+ " is valid by qa and dev")
		return true
	}

	return false
}

func isPullRequestAbleToMerge(pullRequest map[string]interface{}) (bool, int) {
	number := int(pullRequest["number"].(float64))
	numberString := strconv.Itoa(number)

	isHotfix := isHotfix(pullRequest, numberString)

	if !isHotfix {
		return false,number
	}

	isReviewValid := isReviewValid(numberString)

	if !isReviewValid {
		return false,number
	}

	isFilesChangedValid := isFilesChangedValid(numberString)

	if !isFilesChangedValid {
		return false,number
	}

	return true,number
}

func mergePullRequest(number int) {
	numberString := strconv.Itoa(number)

	fmt.Println("Pull Request: "+numberString+" is ok to review")

	client := &http.Client{}

	req, _ := http.NewRequest("PUT", mainUrl+"/"+numberString+"/merge", nil)
	req.Header.Add("Authorization", "token f66d639702c0a0d792251fb7e3c06386d54df629")
	client.Do(req)
}

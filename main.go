package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
)

func main() {

	// For loop that check all the websites for updates, make sure to increase the "i" variable when adding more websites to the list
	for i := 0; i < 3; i++ {

		// Variables needed, not all will be used everytimes depends on the function you are chosing to use
		var websiteLink string
		var websiteFile string
		var website string
		var websiteFileOld string
		var websiteFileNew string
		var stringToFind string
		var modifyRequirement bool

		// Poocoin.app
		if i == 0 {
			website = "Poocoin.app"
			websiteLink = "https://poocoin.app/whitelist1-tokens.json"
			websiteFile = "./last_response_poocoin.txt"
			checkUpdates(website, websiteLink, websiteFile)
		}

		//Coingecko
		if i == 1 {
			website = "Coingecko"
			websiteLink = "https://www.coingecko.com/en/coins/recently_added?page=1"
			websiteFileOld = "./last_response_coingecko_old.txt"
			websiteFileNew = "./last_response_coingecko_new.txt"
			stringToFind = "py-0 coin-name"
			modifyRequirement = false
			checkUpdatesSpecificLine(website, websiteLink, websiteFileOld, websiteFileNew, stringToFind, modifyRequirement)

		}

		//Coin Market Cap
		if i == 2 {
			website = "Coin Market Cap"
			websiteLink = "https://coinmarketcap.com/new/"
			websiteFileOld = "./last_response_cmc_old.txt"
			websiteFileNew = "./last_response_cmc_new.txt"
			stringToFind = "p font-weight=\"semibold\""
			modifyRequirement = true
			checkUpdatesSpecificLine(website, websiteLink, websiteFileOld, websiteFileNew, stringToFind, modifyRequirement)

			i = -1
		}

	}
}

// This function is required for when website list all of the html code in a single line, this will
// spread the code out into multiple lines for the scanner to find what its looking for, it does this by
// seperating the code at every "<" symbol found.
func lineSplitter(file string) {

	var err error

	//opens the text file
	fileText, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	// splits the text by every "<" symbol, setting all the html coded to new lines
	result := strings.Split(string(fileText), "<")

	// compliles all the new lines and saves it to a slice
	stringSlices := strings.Join(result, "\n")

	//fmt.Println(stringSlices) testing code

	// writes it to the file
	ioutil.WriteFile(file, []byte(stringSlices), 0644)

}

// This function is used to find the specific line of code from the website that allows you to know when a change occurs
// to the webpage, input in the specific string you are looking for and then the file you are searching for it in
// then it will save all of the specific strings it finds to the file
func readLine(stringToFind string, file string) (line string) {

	var err error
	var substrFound []string

	// Opens the file
	scan, err := os.Open(file)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	// Scans the file to find the specific string you are looking for that contains the token name
	sc := bufio.NewScanner(scan)
	for sc.Scan() {
		if strings.Contains(sc.Text(), stringToFind) {
			//fmt.Println("Found the line: " + sc.Text()) code for testing purposes
			substrFound = append(substrFound, sc.Text())
			// return sc.Text()
		}
	}
	// Adds the found substrings to the slice then saves the slice to the file
	stringSlices := strings.Join(substrFound, "\n")
	ioutil.WriteFile(file, []byte(stringSlices), 0644)

	// Returns the found substrings and none of the other text
	return stringSlices
}

// Function used for websites that require searching for the line of code that changes when a new coin is added, these require 2 website files, one old one and one new one for the
// function to compare the two together, check for any changes, then notify via email if there is a change
func checkUpdatesSpecificLine(website string, websiteLink string, websiteFileOld string, websiteFileNew string, stringToFind string, modifyRequirement bool) {

	fmt.Println("fetching updates for " + website)

	// Copies the code from the website
	r, err := http.Get(websiteLink)
	if err != nil {
		// handle error
	}
	defer r.Body.Close()

	// Reads the code from the website
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Writes the code to the new file from the website
	response := string(body)
	ioutil.WriteFile(websiteFileNew, []byte(response), 0644)

	// this is required for when website list all of the html code in a single line, this will
	// spread the code out into multiple lines for the scanner to find what its looking for
	if modifyRequirement {
		lineSplitter(websiteFileNew)
		lineSplitter(websiteFileOld)
	}

	// Compares the newly written code to the old one. By using the readline function it goes through the code until it find the first listed coin and then compares the newly
	// saved file to the old one in the same exact spot, allowing it to notice if there is a change. If there is a change it will send a email and update the old file.
	if readLine(stringToFind, websiteFileNew) != readLine(stringToFind, websiteFileOld) {
		ioutil.WriteFile(websiteFileOld, []byte(readLine(stringToFind, websiteFileNew)), 0644)
		//email
		fmt.Println("change")
		sendEmail(website)
	} else {
		// do nothing
		fmt.Println("no change")
	}

	// Checks for changes every 15 seconds
	time.Sleep(time.Minute / 8)
}

// Function to check for updates used when the coin lists are easy accessable and can just compare them for updates. See checkupdatesspecificline for rundown on how the funtion
// operates (not exactly the same buy very simliar)
func checkUpdates(website string, websiteLink string, websiteFile string) {

	fmt.Println("fetching updates for " + website)

	r, err := http.Get(websiteLink)
	if err != nil {
		// handle error
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	response := string(body)

	lastResponse, err := ioutil.ReadFile(websiteFile)
	if err != nil {
		log.Fatal(err)
	}

	if response != string(lastResponse) {
		ioutil.WriteFile(websiteFile, []byte(response), 0644)
		//email
		fmt.Println("change")
		sendEmail(website)
	} else {
		// do nothing
		fmt.Println("no change")
	}

	time.Sleep(time.Minute / 8)
}

// Sends and email when a change occurs
func sendEmail(websiteName string) {
	email := "talos.tester.1@gmail.com"
	pass := "Tester123!"
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	message := []byte(websiteName + " updated their list!")

	// add recipient emails here
	to := []string{
		"talos.tester.1@gmail.com",
		//"7rossilli7@gmail.com",
		//"butchjacob23@gmail.com",
		//"Ccarbasho@gmail.com",
		//"Tuttle.isaiah@gmail.com",
		//"rrossilli55@gmail.com",
		//"Mrmonihan@hotmail.com",
		//"Mrmonihan1@gmail.com",
	}

	auth := smtp.PlainAuth("", email, pass, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, email, to, message)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Email Sent Successfully!")

}

package hscan

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"os"
)

//==========================================================================\\

var shalookup map[string]string
var md5lookup map[string]string

func GuessSingle(sourceHash string, filename string) string {

	f, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	returnVal := "Not Found"

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		password := scanner.Text()

		length := len(sourceHash)

		if length == 32 {
			hash := fmt.Sprintf("%x", md5.Sum([]byte(password)))
			if hash == sourceHash {
				//fmt.Printf("[+] Password found (MD5): %s\n", password)
				returnVal = password
			}
		} else if length == 64 {
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
			if hash == sourceHash {
				//fmt.Printf("[+] Password found (SHA-256): %s\n", password)
				returnVal = password
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalln(err)
	}
	return returnVal
}

func GenHashMaps(filename string) (int, int) {

	f, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		password := scanner.Text()

		go GetSHA(password)

		go GetMD5(password)
	}

	return len(md5lookup), len(shalookup)
	//TODO
	//itterate through a file (look in the guessSingle function above)
	//rather than check for equality add each hash:passwd entry to a map SHA and MD5 where the key = hash and the value = password
	//TODO at the very least use go subroutines to generate the sha and md5 hashes at the same time
	//OPTIONAL -- Can you use workers to make this even faster

	//TODO create a test in hscan_test.go so that you can time the performance of your implementation
	//Test and record the time it takes to scan to generate these Maps
	// 1. With and without using go subroutines
	// 2. Compute the time per password (hint the number of passwords for each file is listed on the site...)
}

func GetSHA(hash string) (string, error) {
	password, ok := shalookup[hash]
	if ok {
		return password, nil

	} else {

		return "", errors.New("password does not exist")

	}
}

func GetMD5(hash string) (string, error) {
	password, ok := md5lookup[hash]
	if ok {
		return password, nil

	} else {

		return "", errors.New("password does not exist")

	}
}

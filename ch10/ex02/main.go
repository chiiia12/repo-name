package main

import (
	"fmt"
	"./driver"
	_ "./zip"
	_ "./tar"
)

const OUTPUT_DIR = "./out"

func main() {
	//unzip, err := driver.OpenUnArchiver("zip")
	//if err != nil {
	//	fmt.Println("OpenUnArchiver has error.", err)
	//}
	//unzip.UnArchive("./sample.zip", OUTPUT_DIR)

	untar, err := driver.OpenUnArchiver("tar")
	if err != nil {
		fmt.Println("OpenUnArchiver has error.", err)
	}
	untar.UnArchive("./sample.tar", OUTPUT_DIR)
}

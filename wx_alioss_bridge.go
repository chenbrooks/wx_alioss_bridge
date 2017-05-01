package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	_ "github.com/go-sql-driver/mysql"
)


var oss_domain = "oss-cn-shenzhen.aliyuncs.com";

func main() {
	webServer()
}

/**
	监听本机的9527端口，如果有请求就处理
 */
func webServer() {
	fmt.Println("start server")
	http.HandleFunc("/aliyun/wx", wx2Oss)
	http.HandleFunc("/aliyun/xiumi", xiumi2Oss)
	http.HandleFunc("/aliyun/local", local2Oss)

	err := http.ListenAndServe(":9527", nil);
	if err != nil{
		log.Fatal(err)
	}
}

func local2Oss(writer http.ResponseWriter, request *http.Request){

	request.ParseForm()
	param_imageDir, found1 := request.Form["imagedir"]
	param_bucketName, found3 := request.Form["bucketname"];
	param_accessKeyID, found4 := request.Form["keyid"];
	param_accessKeySecret, found5 := request.Form["keysecret"];
	if !found1 {
		fmt.Fprint(writer, "please provide image info: image dir")
		return
	}

	if !(found3 && found4 && found5){
		fmt.Fprint(writer, "please provide oss info: bucketname or keyid or key")
		return
	}
	filename := downloadLocalImage(param_imageDir[0])

	fmt.Println(filename);
	fullAddress := uploadAliyunImage(filename, param_bucketName[0], param_accessKeyID[0], param_accessKeySecret[0] )
	fmt.Print( fullAddress)
	fmt.Fprint(writer, fullAddress)
}


func xiumi2Oss(writer http.ResponseWriter, request *http.Request){
	request.ParseForm()
	param_imageURL, found1 := request.Form["imageurl"]
	param_bucketName, found3 := request.Form["bucketname"];
	param_accessKeyID, found4 := request.Form["keyid"];
	param_accessKeySecret, found5 := request.Form["keysecret"];
	if !found1 {
		fmt.Fprint(writer, "please provide image info: image url")
		return
	}

	if !(found3 && found4 && found5){
		fmt.Fprint(writer, "please provide oss info: bucketname or keyid or key")
		return
	}
	filename := downloadXiuMiImage(param_imageURL[0])

	fmt.Println(filename);
	fullAddress := uploadAliyunImage(filename, param_bucketName[0], param_accessKeyID[0], param_accessKeySecret[0] )
	fmt.Print( fullAddress)
	fmt.Fprint(writer, fullAddress)
}

func wx2Oss(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	param_accessToken, found1 := request.Form["accesstoken"]
	param_imageID, found2:= request.Form["imageid"]
	param_bucketName, found3 := request.Form["bucketname"];
	param_accessKeyID, found4 := request.Form["keyid"];
	param_accessKeySecret, found5 := request.Form["keysecret"];

	fmt.Println(param_bucketName);
	if !(found1 && found2){
		fmt.Fprint(writer, "please provide wx info: accesstoken or imageid")
		return
	}

	if !(found3 && found4 && found5){
		fmt.Fprint(writer, "please provide oss info: bucketname or keyid or key")
		return
	}

	filename := downloadWXImage(param_accessToken[0], param_imageID[0])

	fmt.Println(filename);
	fullAddress := uploadAliyunImage(filename, param_bucketName[0], param_accessKeyID[0], param_accessKeySecret[0] )
	fmt.Print( fullAddress);
	fmt.Fprint(writer, fullAddress)

}





// http://idcars-test.oss-cn-shenzhen.aliyuncs.com/
func getFullUrl(filename string, bucketname string) string {
	return "http://" + bucketname + "."+oss_domain+ "/" + filename;
}


func downloadWXImage(accessToken string, wxImageID string) string {
	rawURL := "https://api.weixin.qq.com/cgi-bin/media/get?access_token="
	url := rawURL + accessToken + "&media_id=" + wxImageID
	return downloadFile(url);
}

func downloadXiuMiImage(imageUrl string) string{
	return downloadFile(imageUrl)
}

func downloadLocalImage(imageDir string) string{

	filename := getFileName()
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
		fmt.Println(err)
	}
	src, _:= os.Open(imageDir)
	_, err = io.Copy(file,	src)
	if err != nil {
		log.Fatal(err)
		fmt.Println(err)
	}
	file.Close()
	fmt.Println("download success")
	return filename;
}

func downloadFile(url string) string  {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		fmt.Println(err)
	}
	defer res.Body.Close()
	filename := getFileName()
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
		fmt.Println(err)
	}
	_, err = io.Copy(file, res.Body)
	if err != nil {
		log.Fatal(err)
		fmt.Println(err)
	}
	file.Close()
	fmt.Println("download success")
	return filename;
}


func uploadAliyunImage(filename string, bucketName string, keyID string, keySecret string) string {
	aliyunFile := strings.Replace(filename, "/tmp/", "", 1)
	//if(keyID == ""){
	client, err := oss.New(oss_domain, "LTAIZ9aDNO2SCUex", "48vj1cOVe1RRBTxgNiEMujZI3mSE35")
	//}

	//client, err := oss.New(oss_domain, keyID, keySecret)
	if err != nil {
		log.Fatal(err)
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		log.Fatal(err)
	}
	err = bucket.PutObjectFromFile(aliyunFile, filename)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("upload success")
	return getFullUrl(aliyunFile, bucketName)
}

func getFileName() string {
	layout := "200601/02"
	dir := "/tmp/" + (time.Now().Format(layout))
	os.MkdirAll(dir, 0755)
	return dir + "/" + getMD5("salt") + ".jpg"
}

func getMD5(imageID string) string {
	timeInt := imageID + strconv.Itoa(time.Now().Nanosecond())
	h := md5.New()
	h.Write([]byte(timeInt))
	return hex.EncodeToString(h.Sum(nil))
}

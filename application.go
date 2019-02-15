package main

import (
    "os"
    "io"
    "log"
    "time"
    //"fmt"
    "net/http"
    //"io/ioutil"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
)

func handler(writer http.ResponseWriter, request *http.Request) {
  //os.Setenv below is only needed when troubleshooting the app locally
  //os.Setenv("AWS_REGION", "us-west-2")
  //os.Setenv("ARTIFACT_BUCKET", "my-artifact-bucket-name")

  // This block of code is commented out so that we don't need to maintain a written logfile
  // it can be uncommented out if we want an actual log file
  //logFile, err := os.OpenFile("main.log", os.O_CREATE | os.O_APPEND | os.O_RDWR, 0666)
  //if err != nil {
  //  panic(err)
  //}
  //log.SetOutput(logFile)

  // Use MultiWriter to allow both console log and writing to logfile simultaneously
  // Swap out below for Single Stdout OR MultiWrite Stdout and Log to file
  //mw := io.MultiWriter(os.Stdout, logFile)
  mw := io.MultiWriter(os.Stdout)
  log.SetOutput(mw)

  //Assuming the Elastic IAM role (or your AWS creds through saml auth)
  //This creates a session with AWS SDK
  sess, err := session.NewSession(&aws.Config{Region: aws.String(os.Getenv("AWS_REGION"))},)
  svc := s3.New(sess)

  //We will pre-sign the URL while passing in the ARTIFACT_BUCKET environment variable
  //and the URL path will translate to the path to the S3 object
  req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
      Bucket: aws.String(os.Getenv("ARTIFACT_BUCKET")),
      Key: aws.String(request.URL.Path[1:]),
  })

  //TODO:  We should validate that the object actually exists first before we sign the URL
  //It is possible to pre-sign an object that does not exist
  //Even though it can be considered user error, we should be able to check against it

  urlStr, err := req.Presign(3 * time.Minute)  // the signed URL has an expiry of 3 minutes

  if err != nil {
      log.Println("Failed to sign request", err)
  }

  //Redirect to the pre-signed URL: web browsers will auto-download
  //while "curl -O -J -L" will follow the redirect and download to the current directory
  log.Printf("Redirect to: %s", urlStr)
  http.Redirect(writer, request, urlStr, 301)
  return
}

func main() {
  http.HandleFunc("/", handler)
  http.ListenAndServe(":8080", nil)  //webapp listens on port 8080
}

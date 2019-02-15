# Simple Artifact Store - private, encrypted S3 bucket
This project was created as a proof-of-concept. It is a Golang web application designed to run in AWS Elastic Beanstalk.  The web application will silently output and redirect pre-signed URLs that allows third parties temporary read access (s3:GetObject) to an S3 object.  These pre-signed URLs have a specific time-to-live and will expire in a specified amount of time.  Typically, sharing data in S3 buckets either requires much adminstrator overhead by configuring bucket policies, and constant modifications to IAM roles to allow cross-account access (especially if you need to allow more end-users or business units as an on-going effort). The main advantage with this application is that there is no need for these constant administrative tasks.  In fact, this allows third parties to perform get operations on a private S3 bucket without them even needing an AWS account.  Please keep in mind that this is a proof-of-concept and more security measures should be implemented.  I will try to list out all the advantages and disadvantages.

### Advantages
* The S3 bucket created is a private bucket (not accessible to the public directly), and server-side encrypted (AES-256).  (I have not tried KMS.)
* You do not need to manage cross-account assumed roles in IAM
* You do not need to manage any bucket policies
* You manage one policy in AWS (attached to the Elastic Beanstalk EC2 role), and can fine-tune that policy by allowing/denying specific folders under the S3 bucket
* The expiration for the pre-signed URL can be configured to any value, seconds, if needed.  Any subsequent requests with the same pre-signed URL after the expiry time has passed will result in "Access Denied" messages
* Authentication can be added for more security including, but not limited to, basic authentication, OAuth2 bearer tokens, JWT

### Disadvantages
* Currently, this POC does not secure the web application in any way (though there is security through obscurity as clients need to know the exact path/to/s3object/filename)
* Due to potential security issues, I would only advise to try this only if the contents allowed through the s3:GetObject policy do not contain any proprietary information.  Or if you were doing a POC of a POC.

### Pre-requisites
* Create a private AES-256 encrypted S3 bucket. Thoughout my examples in this README, I will be the bucket name `artifact-store-test` as an example. Let's also assume that I have created this bucket in `us-west-2`
* Create an IAM `s3:GetObject` policy to be used later. As an example, I have created the following:
```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "GetOjbectFolder1",
            "Effect": "Allow",
            "Action": "s3:GetObject",
            "Resource": "arn:aws:s3:::artifact-store-test/folder1*"
        }
    ]
}
```

In the example above, this will essentially allow clients to perform Get operations only under `folder1`. Clients would not be able to download objects from `folder2`.  These are just simple examples. To be more accurate, this policy would also allow access to folder11, and folder12, if those existed.

### Creating the zip file for Elastic Beanstalk
Run `create_zip.sh`.  This should create `uploadThis.zip` in the same directory.

### Create the Elastic Beanstalk app
1. In AWS console, click on `Services` then `Elastic Beanstalk`
2. Create a new `Application name` (this can be anything you want)
3. For `Platform`, choose Preconfigured / Go
4. Choose to `Upload your code` and select your local file `uploadThis.zip`
5. Click on `Configure more options`
6. Click on `Modify` under the Software section
7. Add the following environment properties for the application:

| __Name__ | __Value__ |
|-------------|------------|
| ARTIFACT_BUCKET    | artifact-store-test     |
| AWS_REGION         | us-west-2 |
8. In the above example, AWS_REGION is `us-west-2` because `artifact-store-test` was created in `us-west-2`. Also note that the value for `ARTIFACT_BUCKET` is the friendly bucket name and not the full arn.
9. Ensure that the security group attached to the Elastic Beanstalk EC2 has the following inbound rules (the webapp listens on 8080):

| __Type__ | __Protocol__ | __Port_Range__ | __Source__ |
|-------------|------------|------------|------------|
| Custom TCP Rule | TCP | 8080 | 0.0.0.0/0
| Custom TCP Rule | TCP | 8080 | ::/0
10. Assuming you are using the default `aws-elasticbeanstalk-ec2-role` instance profile setting for Elastic Beanstalk, go into IAM, select the `aws-elasticbeanstalk-ec2-role` role and attach the policy you created from the pre-requisites section.

### Populated files in S3
To continue my example, my S3 bucket structure looks like this:
```
artifact-store-test
	└── folder1
	    ├── folder1file1.txt  
	    ├── folder1file2.txt
	└── folder2
	    ├── folder2file1.txt
	    ├── folder2file2.txt
```	    

### Example usage
1. curl -O -J -L http://IP.OF.ElasticBeanstalk.EC2:8080/folder1/folder1file1.txt
2. curl -O -J -L http://random.elasticbeanstalkurl.elasticbeanstalk.com:8080/folder1/folder1file2.txt
3. curl -O -J -L http://IP.OF.ElasticBeanstalk.EC2:8080/folder2/folder2file1.txt
4. curl -O -J -L http://IP.OF.ElasticBeanstalk.EC2:8080/folder1/filedoesntexist.txt

The curl commands will follow the redirected URL which is a presigned URL and automatically downloads the file to the current directory. `Example #1` will download `folder1file1.txt` as our IAM policy allows us to do so.  `Example #2` demonstrates that you can use either the ElasticBeanstalk URL, or the public IP of the ElasticBeanstalk EC2 instance (again, be sure that port 8080 is allowed in the security group).  `Example #3` tries to download a file that exists in `folder2` which is not allowed through our IAM policy. The client will be unable to download the file and will receive an `Access Denied` message. `Example #4` shows that the client tries to download a file that does not exist in a folder that he has permissions to.  The actual filename does get created on the local system, but it will result in an `Access Denied` message (in the contents of the file).

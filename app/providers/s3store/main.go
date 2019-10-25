package s3store

import (
	"bytes"
	"encoding/base64"
	"path/filepath"

	"encoding/csv"
	"errors"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type s3res struct {
	Name   string
	Reader io.Reader
	CsvReader string
}

type S3ImageProvider struct {
	bucket string
	prefix string
	id     string
	secret string
	region string
	hit    chan *s3res
}

func NewS3ImageProvider(id, secret, region, bucket, prefix string) *S3ImageProvider {

	c := aws.NewConfig().WithCredentials(credentials.NewStaticCredentials(id, secret, ""))
	sess, err := session.NewSession(c)
	if err != nil {
		log.Println(err)
	}
	svc := s3.New(sess, aws.NewConfig().WithRegion(region))

	sss := &S3ImageProvider{
		id:     id,
		secret: secret,
		region: region,
		bucket: bucket,
		prefix: prefix,
		hit:    make(chan *s3res),
	}
	go sss.fetch(svc)
	return sss
}

func (sss *S3ImageProvider) GetImage() (string, string, map[int][]string, error) {
	Records := make(map[int][]string)
	if i, ok := <-sss.hit; ok {
		//Parsing the String into []string to make annotation later
		if i.CsvReader != "" {
			reader := csv.NewReader(strings.NewReader(i.CsvReader))
			counter := 0
			for {
				record, err := reader.Read()
				counter += 1
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Println(err)
				}

				Records[counter] =record

				log.Println("len of record")
				log.Println(len(record))
			}
		}
		// copy stream
		im, _, err := image.Decode(i.Reader)
		if err != nil {
			log.Println(err)
		}
		var buff bytes.Buffer

		png.Encode(&buff, im)
		return i.Name, "data:image/png;base64," + base64.StdEncoding.EncodeToString(buff.Bytes()), Records, nil
	}
	return "", "", Records,errors.New("No new file")
}

func (sss *S3ImageProvider) AddImage(name, url string) {
	var b []byte

	buff := bytes.NewBuffer(b)
	buff.WriteString(url)

	sss.hit <- &s3res{
		Name:   name,
		Reader: buff,
	}
}

func (sss *S3ImageProvider) fetch(svc *s3.S3) {
	buck := &s3.ListObjectsV2Input{}
	buck.SetBucket(sss.bucket)
	buck.SetPrefix(sss.prefix + "/")
	buck.SetDelimiter("/")

	sss.listThat(svc, buck)
}

func (sss *S3ImageProvider) listThat(svc *s3.S3, buck *s3.ListObjectsV2Input) {
	log.Println("gonne search for img and csv")
	prefixes := []string{}
	page := 0
	var csvString string
	err := svc.ListObjectsV2Pages(buck, func(p *s3.ListObjectsV2Output, lastPage bool) bool {
		page++
		//var newBox annotation.Box
		for _, cc := range p.Contents {
			isImage := false
			for _, ext := range []string{".jpg", ".jpeg", ".png"} {
				log.Println("finding...")
				k := strings.ToLower(*cc.Key)
				if strings.HasSuffix(k, ext) {

					isImage = true
				}
			}
			if !isImage {
				// Switch to next object in bucket
				continue
			}

			// Getting image from S3
			in := s3.GetObjectInput{
				Bucket: buck.Bucket,
				Key:    cc.Key,
			}
			res, err := svc.GetObject(&in)
			if err != nil {
				log.Println("S3 ERR 1", err)
			}

			var filename = *cc.Key
			var extension = filepath.Ext(filename)
			var CsvName = filename[0:len(filename)-len(extension)] + ".csv"


			csvfile := s3.GetObjectInput{
				Bucket: buck.Bucket,
				Key:    &CsvName,
			}
			rescsv, errcsv := svc.GetObject(&csvfile)
			if errcsv != nil {
				log.Println("S3 ERR 1", errcsv)
				csvString = ""
			}else {
				buf := new(bytes.Buffer)
				buf.ReadFrom(rescsv.Body)
				csvString = buf.String()
				rescsv.Body.Close()
			}

			// AddImage...
			sss.hit <- &s3res{
				Name:   *cc.Key,
				Reader: res.Body,
				CsvReader: csvString,
			}
		}
		for _, cp := range p.CommonPrefixes {
			prefixes = append(prefixes, *cp.Prefix)
		}
		return lastPage
	})
	if err != nil {
		log.Println("S3 failed", err)
	}
	for _, p := range prefixes {
		b := &s3.ListObjectsV2Input{}
		b.SetBucket(*buck.Bucket)
		b.SetPrefix(p)
		b.SetDelimiter("/")
		sss.listThat(svc, b)
	}
}

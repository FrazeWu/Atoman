package storage

import (
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"gorm.io/gorm"

	"amb/internal/model"
)

// InitS3Client initializes and returns an S3 client configured for Oracle Object Storage
func InitS3Client() (*s3.S3, error) {
	// AWS S3 session (compatible with Oracle Object Storage)
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(os.Getenv("S3_REGION")),
		Endpoint:         aws.String(os.Getenv("S3_ENDPOINT")),
		Credentials:      credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
		S3ForcePathStyle: aws.Bool(true), // Required for Oracle Object Storage
	})
	if err != nil {
		return nil, err
	}

	s3Client := s3.New(sess)
	return s3Client, nil
}

// ValidateS3Connection tests the S3 connection and permissions
func ValidateS3Connection(s3Client *s3.S3) error {
	bucket := os.Getenv("S3_BUCKET")

	// Test bucket access
	_, err := s3Client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		log.Printf("S3 validation failed: %v", err)
		return err
	}

	log.Printf("S3 connection validated successfully for bucket: %s", bucket)
	return nil
}

// DeleteS3Object deletes an object from the configured S3 bucket
func DeleteS3Object(s3Client *s3.S3, key string) error {
	bucket := os.Getenv("S3_BUCKET")
	_, err := s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Printf("Failed to delete S3 object %s from bucket %s: %v", key, bucket, err)
		return err
	}
	log.Printf("Successfully deleted S3 object: %s from bucket: %s", key, bucket)
	return nil
}

// DeleteAlbumAndS3Objects deletes an album record and its associated S3 cover image
func DeleteAlbumAndS3Objects(db *gorm.DB, s3Client *s3.S3, album *model.Album) error {
	if album.CoverURL != "" {
		coverKey := strings.TrimPrefix(album.CoverURL, os.Getenv("S3_URL_PREFIX")+"/")
		if err := DeleteS3Object(s3Client, coverKey); err != nil {
			log.Printf("Failed to delete cover file %s from S3: %v", coverKey, err)
			// Don't return error here, try to delete from DB anyway
		}
	}

	if err := db.Delete(&album).Error; err != nil {
		log.Printf("Failed to delete album %d from DB: %v", album.ID, err)
		return err
	}

	return nil
}

// SaveFileLocally saves an uploaded file to local uploads directory
// Returns the local file path and URL for database storage
func SaveFileLocally(file interface{}, filename, artist, album string) (string, string, error) {
	// Import required for file operations
	var reader interface {
		Read([]byte) (int, error)
	}

	// Type assertion for file reader
	if f, ok := file.(interface{ Read([]byte) (int, error) }); ok {
		reader = f
	} else {
		return "", "", os.ErrInvalid
	}

	// Create safe directory names
	safeArtist := strings.ReplaceAll(artist, "/", "-")
	if safeArtist == "" {
		safeArtist = "Unknown Artist"
	}
	safeAlbum := strings.ReplaceAll(album, "/", "-")
	if safeAlbum == "" {
		safeAlbum = "Unknown Album"
	}

	// Create directory structure: uploads/music/Artist/Album/
	dir := "uploads/music/" + safeArtist + "/" + safeAlbum
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", err
	}

	// Create local file path
	localPath := dir + "/" + filename

	// Create file
	dst, err := os.Create(localPath)
	if err != nil {
		return "", "", err
	}
	defer dst.Close()

	// Copy uploaded file to local file
	buffer := make([]byte, 32*1024)
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			if _, writeErr := dst.Write(buffer[:n]); writeErr != nil {
				return "", "", writeErr
			}
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", "", err
		}
	}

	// Generate local URL (relative path for serving)
	localURL := "/uploads/music/" + safeArtist + "/" + safeAlbum + "/" + filename

	log.Printf("File saved locally: %s", localPath)
	return localPath, localURL, nil
}

// UploadLocalFileToS3 uploads a local file to S3 and returns the S3 URL
func UploadLocalFileToS3(s3Client *s3.S3, localPath string) (string, error) {
	// Open local file
	file, err := os.Open(localPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Extract S3 key from local path
	// uploads/music/Artist/Album/file.mp3 -> music/Artist/Album/file.mp3
	s3Key := strings.TrimPrefix(localPath, "uploads/")
	s3Key = strings.ReplaceAll(s3Key, "\\", "/") // Windows compatibility

	// Upload to S3
	bucket := os.Getenv("S3_BUCKET")
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(s3Key),
		Body:   file,
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		return "", err
	}

	// Generate S3 URL
	s3URL := os.Getenv("S3_URL_PREFIX") + "/" + s3Key
	log.Printf("File uploaded to S3: %s -> %s", localPath, s3URL)

	return s3URL, nil
}

// DeleteLocalFile deletes a file from local uploads directory
func DeleteLocalFile(localPath string) error {
	if localPath == "" {
		return nil
	}

	// Only delete files in uploads directory for safety
	if !strings.HasPrefix(localPath, "uploads/") {
		log.Printf("Skipping deletion of non-upload file: %s", localPath)
		return nil
	}

	if err := os.Remove(localPath); err != nil {
		log.Printf("Failed to delete local file %s: %v", localPath, err)
		return err
	}

	log.Printf("Successfully deleted local file: %s", localPath)
	return nil
}

// GetLocalPathFromURL extracts local file path from URL
// /uploads/music/Artist/Album/file.mp3 -> uploads/music/Artist/Album/file.mp3
func GetLocalPathFromURL(url string) string {
	if strings.HasPrefix(url, "/uploads/") {
		return strings.TrimPrefix(url, "/")
	}
	return ""
}

// DeleteSongAndS3Objects deletes a song record and its associated files
func DeleteSongAndS3Objects(db *gorm.DB, s3Client *s3.S3, song *model.Song) error {
	if song.AudioURL != "" {
		if song.AudioSource == "local" {
			localPath := GetLocalPathFromURL(song.AudioURL)
			if localPath != "" {
				DeleteLocalFile(localPath)
			}
		} else if song.AudioSource == "s3" {
			audioKey := strings.TrimPrefix(song.AudioURL, os.Getenv("S3_URL_PREFIX")+"/")
			if err := DeleteS3Object(s3Client, audioKey); err != nil {
				log.Printf("Failed to delete audio file %s from S3: %v", audioKey, err)
			}
		}
	}

	if song.CoverURL != "" && (song.Album == nil || song.CoverURL != song.Album.CoverURL) {
		if song.CoverSource == "local" {
			localPath := GetLocalPathFromURL(song.CoverURL)
			if localPath != "" {
				DeleteLocalFile(localPath)
			}
		} else if song.CoverSource == "s3" {
			coverKey := strings.TrimPrefix(song.CoverURL, os.Getenv("S3_URL_PREFIX")+"/")
			if err := DeleteS3Object(s3Client, coverKey); err != nil {
				log.Printf("Failed to delete cover file %s from S3: %v", coverKey, err)
			}
		}
	}

	if err := db.Delete(&song).Error; err != nil {
		log.Printf("Failed to delete song %d from DB: %v", song.ID, err)
		return err
	}

	return nil
}

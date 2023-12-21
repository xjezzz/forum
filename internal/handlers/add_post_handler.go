package handlers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"forum-project/entities"
	"forum-project/internal"
	logg "forum-project/internal/forum_logger"
	rl "forum-project/internal/request_limiter"
	"forum-project/internal/tmpl"
	"image"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	openai "github.com/sashabaranov/go-openai"

	cloud "cloud.google.com/go/storage"
	vision "cloud.google.com/go/vision/apiv1"
	"cloud.google.com/go/vision/v2/apiv1/visionpb"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"google.golang.org/api/option"
)

type Result struct {
	Categories     map[string]bool    `json:"categories"`
	CategoryScores map[string]float64 `json:"category_scores"`
	Flagged        bool               `json:"flagged"`
}

type Response struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Results []Result `json:"results"`
}

func AddPostHandler(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
	// TODO add limit to add post
	err := limiter.Limit()
	if err != nil {
		logg.ErrorLog.Println(err)
		http.Error(w, "Too many requests", http.StatusBadRequest)
		return
	}
	userId, isLogged := internal.UserIsLogged(r, storage)
	if !isLogged {
		logg.ErrorLog.Println("Unauthorized users cant add posts")
		ErrorHandler(w, http.StatusUnauthorized, "Unauthorized users cant add posts")
		return
	}
	switch r.Method {
	case http.MethodGet:
		{
			tags, err := getAllTags(storage)
			if err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusInternalServerError, "Error with getting tags from database")
				return
			}
			tmpl.RenderTemplate(w, "create_post_page.html", tags)
		}
	case http.MethodPost:
		{
			post := entities.Post{
				UserId: userId,
			}
			openAIApiKey := "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"
			textFromPost := r.FormValue("post")
			title := r.FormValue("title")
			c := openai.NewClient(openAIApiKey)
			ctx := context.Background()

			req := openai.ModerationRequest{
				Input: textFromPost + title,
			}
			resp, err := c.Moderations(ctx, req)
			if err != nil {
				fmt.Println(err)
			}
			b, err := json.Marshal(resp)
			if err != nil {
				fmt.Println(err)
			}
			var response Response

			if err := json.Unmarshal(b, &response); err != nil {
				fmt.Println("Ошибка при декодировании JSON:", err)
				return
			}
			flagged := response.Results[0].Flagged
			if flagged {
				logg.ErrorLog.Printf("User with ID: %v trying to create unsafe post", userId)
				ErrorHandler(w, http.StatusBadRequest, "You cannot create a post like this!\n")
				return
			}
			file, handler, err := r.FormFile("myFile")
			if err != nil && !errors.Is(err, http.ErrMissingFile) {
				logg.InfoLog.Println("Size of file minimum 20 MB!")
				ErrorHandler(w, http.StatusBadRequest, "Invalid file")
				return
			}

			if err == nil {
				defer file.Close()
				post.ImageName, err = validImg(file, handler.Filename)
				if err != nil {
					logg.ErrorLog.Println(err)
					ErrorHandler(w, http.StatusBadRequest, "Invalid file")
					return
				}
				var storageClient *cloud.Client
				client, err := vision.NewImageAnnotatorClient(ctx, option.WithCredentialsFile("winter-quanta-401511-608da83d80f0.json"))
				if err != nil {
					log.Fatalf("Failed to create Vision client: %v", err)
				}

				storageClient, err = cloud.NewClient(ctx, option.WithCredentialsFile("winter-quanta-401511-608da83d80f0.json"))
				if err != nil {
					log.Fatalf("Failed to create Storage client: %v", err)
				}

				bucketName := "forum_backet"
				objectName := post.ImageName
				imageURI := "gs://" + bucketName + "/" + objectName
				bucket := storageClient.Bucket(bucketName)
				obj := bucket.Object(objectName)

				dw := obj.NewWriter(ctx)
				file.Seek(0, 0)
				if _, err := io.Copy(dw, file); err != nil {
					log.Fatalf("Error reading object data: %v", err)
				}

				if err := dw.Close(); err != nil {
					log.Fatalf("Error closing storage writer: %v", err)
				}

				// Создание объекта Image
				image := vision.NewImageFromURI(imageURI)
				// Выполнение анализа изображения
				resp, err := client.DetectSafeSearch(ctx, image, nil)
				if err != nil {
					log.Fatalf("Error detecting safe search: %v", err)
				}
				if resp.GetAdult() == visionpb.Likelihood_VERY_LIKELY {
					r, err := obj.NewReader(ctx)
					if err != nil {
						logg.ErrorLog.Println(err)
						return
					}
					blurredObjectName := "blurred_" + objectName

					inputImage, err := imaging.Decode(r)
					if err != nil {
						return
					}

					blurredImage := imaging.Blur(inputImage, 20.0) // Применить размытие с радиусом 4

					localBlurredFileName := "images/blurred_" + objectName
					localBlurredFile, err := os.Create(localBlurredFileName)
					if err != nil {
						log.Fatalf("Error creating local blurred file: %v", err)
						return
					}
					defer localBlurredFile.Close()
					err = imaging.Encode(localBlurredFile, blurredImage, imaging.JPEG)
					if err != nil {
						return
					}
					post.ImageName = blurredObjectName
					log.Printf("Blurred image uploaded")
				}
			}

			tags := r.Form["tag"]
			post.Body = textFromPost
			post.Title = title
			post.Tags = tags

			err = createPost(&post, storage)
			if err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusInternalServerError, "")
				return
			}
			logg.InfoLog.Println("Post was added")

			http.Redirect(w, r, "/main", http.StatusSeeOther)
		}
	default:
		logg.ErrorLog.Println("Trying open page with incorrect method")
		ErrorHandler(w, http.StatusMethodNotAllowed, "")
		return
	}
}

func applyBlur(inputImage image.Image) image.Image {
	return imaging.Blur(inputImage, 4.0) // Применить размытие с радиусом 4
}

func createPost(post *entities.Post, storage *sql.DB) error {
	record := `INSERT INTO posts(title, body, user_id, img_name) VALUES (?, ?, ?, ?)`
	query, err := storage.Prepare(record)
	if err != nil {
		return err
	}
	result, err := query.Exec(post.Title, post.Body, post.UserId, post.ImageName)
	if err != nil {
		return err
	}
	postId, err := result.LastInsertId()
	if err != nil {
		return err
	}
	record = `INSERT INTO posts_tags(post_id, tag_id) VALUES (?, ?)`
	query, err = storage.Prepare(record)
	if err != nil {
		return err
	}
	for i := range post.Tags {
		_, err = query.Exec(strconv.FormatInt(postId, 10), post.Tags[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func validImg(file multipart.File, fileName string) (string, error) {
	var err error
	uniqueFileName := uuid.New().String()
	hash := sha256.Sum256([]byte(fileName))
	hashedName := hex.EncodeToString(hash[:])
	buff := make([]byte, 512)
	if _, err = file.Read(buff); err != nil {
		return "", err
	}
	switch extension := http.DetectContentType(buff); extension {
	case "image/png":
		fileName = hashedName + uniqueFileName + ".png"
	case "image/gif":
		fileName = hashedName + uniqueFileName + ".gif"
	case "image/jpeg":
		fileName = hashedName + uniqueFileName + ".jpeg"
	default:
		err = errors.New("Incorrect file format")
		return "", err
	}
	outFile, err := os.Create("images/" + fileName)
	if err != nil {
		err = errors.New("Cannot create file")
		return "", err
	}
	defer outFile.Close()
	_, err = outFile.Write(buff)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(outFile, file)
	if err != nil {
		return "", err
	}
	return fileName, nil
}

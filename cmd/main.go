package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/google/go-github/v65/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func main() {
	// 環境変数を読み込む
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// GitHubトークンを環境変数から取得
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN is not set")
	}

	// OAuth2クライアントの作成
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	// GitHubクライアントの作成
	client := github.NewClient(tc)

	// ユーザーのリポジトリを取得
	// ---
	// 特定の単語を用いてフィルタしたいかを選択できるようにしたい
	fmt.Printf("Filter: ")
	var filter string
	fmt.Scanln(&filter)
	var repos []*github.Repository
	// 入力した単語があれば、その単語を含むリポジトリのみを表示する
	if filter != "" {
		options := &github.SearchOptions{
			ListOptions: github.ListOptions{
				Page:    1,
				PerPage: 100,
			},
		}
		// 認証したユーザーのリポジトリを取得
		user, _, err := client.Users.Get(ctx, "")
		if err != nil {
			log.Fatal(err)
		}
		userLogin := user.Login
		query := fmt.Sprintf(
			"%s user:%s",
			filter,
			*userLogin,
		)
		searchResult, _, err := client.Search.Repositories(ctx, query, options)
		if err != nil {
			log.Fatal(err)
		}
		repos = searchResult.Repositories
	} else {
		options := &github.RepositoryListByAuthenticatedUserOptions{
			Visibility: "public",
			ListOptions: github.ListOptions{
				Page:    1,
				PerPage: 100,
			},
		}
		repos, _, err = client.Repositories.ListByAuthenticatedUser(ctx, options)

		if err != nil {
			log.Fatal(err)
		}
	}

	for _, repo := range repos {
		topics := repo.Topics
		if len(topics) == 0 {
			fmt.Printf("Repository: %s\n", *repo.Name)
			fmt.Println("No topics")
			continue
		}
		// Starがついている場合はスキップ
		if *repo.StargazersCount > 0 {
			fmt.Printf("Star: %s\n", *repo.Name)
			continue
		}
		// フォークされたリポジトリの場合はスキップ
		if *repo.ForksCount > 0 {
			fmt.Printf("Fork: %s\n", *repo.Name)
			continue
		}
		fmt.Printf("Repository: %s\n", *repo.Name)
		fmt.Printf("Tags: %v\n", topics)
		if slices.Contains(topics, "public") {
			// パブリックリポジトリでなければ、パブリックリポジトリに変更する
			if !*repo.Private {
				continue
			}
			// 本当にパブリックリポジトリに変更するか確認する
			fmt.Printf("Make public: %s\n", *repo.Name)
			fmt.Printf("Are you sure? (y/n): ")
			var input string
			fmt.Scanln(&input)
			if input != "y" {
				fmt.Println("Canceled")
				continue
			}
			_, _, err := client.Repositories.Edit(ctx, *repo.Owner.Login, *repo.Name, &github.Repository{
				Private: github.Bool(false),
			})
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Made public: %s\n", *repo.URL)
			continue
		}
		if slices.Contains(topics, "private") {
			// プライベートリポジトリでなければ、プライベートリポジトリに変更する
			if *repo.Private {
				continue
			}
			_, _, err := client.Repositories.Edit(ctx, *repo.Owner.Login, *repo.Name, &github.Repository{
				Private: github.Bool(true),
			})
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Made private: %s\n", *repo.Name)
			continue
		}
		if slices.Contains(topics, "delete") {
			// 本当に削除して良いか確認する
			fmt.Printf("Delete: %s\n", *repo.Name)
			fmt.Printf("Are you sure? (y/n): ")
			var input string
			fmt.Scanln(&input)
			if input != "y" {
				fmt.Println("Canceled")
				continue
			}
			// 削除する
			_, err := client.Repositories.Delete(ctx, *repo.Owner.Login, *repo.Name)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Deleted: %s\n", *repo.Name)
			continue
		}
	}
}

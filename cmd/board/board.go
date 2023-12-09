package main

import (
	"context"
	"fmt"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/grpcauth"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/sprites"
	gameserverpb "github.com/c4t-but-s4d/ctfcup-2023-igra/proto/go/gameserver"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"html/template"
	"log"
	"net/http"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"
)

type hostInfo struct {
	Host  string
	Token string
}

var teamToHost = map[string]hostInfo{
	"C4T B4T S4D": {
		Host:  "localhost:8080",
		Token: "CBS:abobus",
	},
	"Team 2": {
		Host:  "localhost:8085",
		Token: "NotCBS:abobus",
	},
}

type server struct {
	db  *bolt.DB
	mng *sprites.Manager
}

func (s *server) renderIndex(w http.ResponseWriter, r *http.Request) {
	itemsCollected := make(map[string][]string)
	if err := s.db.View(func(tx *bolt.Tx) error {
		for teamName := range teamToHost {
			b := tx.Bucket([]byte(teamName))
			if b == nil {
				log.Printf("bucket %s not found", teamName)
				continue
			}
			return b.ForEach(func(k, v []byte) error {
				itemsCollected[teamName] = append(itemsCollected[teamName], string(k))
				return nil
			})
		}
		return nil
	}); err != nil {
		log.Printf("failed to get items collected: %v", err)
	}

	type Team struct {
		TeamName       string
		Points         int
		ItemsCollected []string
	}
	type TeamsData struct {
		Teams []Team
	}

	var teams []Team
	for teamName := range teamToHost {
		teams = append(teams, Team{
			TeamName:       teamName,
			Points:         len(itemsCollected[teamName]),
			ItemsCollected: itemsCollected[teamName],
		})
	}

	sort.Slice(teams, func(i, j int) bool {
		return teams[i].Points > teams[j].Points
	})

	tmpl, err := template.New("teamList").ParseFiles("index.html")
	if err != nil {
		log.Printf("failed to parse template: %v", err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "teamList", TeamsData{Teams: teams}); err != nil {
		log.Printf("failed to execute template: %v", err)
		return
	}
}

func updateInventory(ctx context.Context, db *bolt.DB, teamName string, info hostInfo) error {
	var client gameserverpb.GameServerServiceClient
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	interceptor := grpcauth.NewClientInterceptor(info.Token)
	opts = append(
		opts,
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
	)

	conn, err := grpc.DialContext(ctx, info.Host, opts...)
	if err != nil {
		return fmt.Errorf("dialing to %s: %w", info.Host, err)
	}

	client = gameserverpb.NewGameServerServiceClient(conn)
	ir, err := client.GetInventory(ctx, &gameserverpb.InventoryRequest{})
	if err != nil {
		return fmt.Errorf("getting inventory: %w", err)
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(teamName))
		if err != nil {
			return fmt.Errorf("creating bucket: %w", err)
		}
		for _, item := range ir.GetInventory().GetItems() {
			if err := b.Put([]byte(item.GetName()), []byte(item.GetName())); err != nil {
				return fmt.Errorf("putting item: %w", err)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("updating db: %w", err)
	}

	return nil
}

func main() {
	db, err := bolt.Open("board.db", 0755, nil)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	t := time.NewTicker(time.Second * 5)
	defer t.Stop()
	go func() {
		// Update
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-t.C:
				var wg sync.WaitGroup
				wg.Add(len(teamToHost))
				log.Printf("updating inventory at %v", t)

				for teamName, info := range teamToHost {
					go func(name string, info hostInfo) {
						defer wg.Done()
						if err := updateInventory(ctx, db, name, info); err != nil {
							log.Printf("failed to update inventory for %s: %v", name, err)
						}
					}(teamName, info)
				}

				wg.Wait()
				log.Printf("finished updating inventory at %v", t)
			}
		}
	}()

	s := &server{
		db:  db,
		mng: sprites.NewManager(),
	}
	http.HandleFunc("/", s.renderIndex)
	go func() {
		log.Printf("starting server on :9091\n")
		log.Fatal(http.ListenAndServe(":9091", nil))
	}()

	<-ctx.Done()
}

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	dbpkg "github.com/TinySkillet/DecentralizedP2PStorage/db"
	"github.com/spf13/cobra"
)

func setupCommands() *cobra.Command {
	var (
		listen    string
		dbPath    string
		bootstrap []string
	)

	root := &cobra.Command{Use: "p2p", Short: "Decentralized P2P storage node"}
	root.PersistentFlags().StringVar(&dbPath, "db", "p2p.db", "sqlite database path")

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Run a node",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := dbpkg.Open(dbPath)
			if err != nil {
				return err
			}
			defer d.Close()
			if err := d.Migrate(context.Background()); err != nil {
				return err
			}
			keyBytes, err := loadOrInitKey(d)
			if err != nil {
				return err
			}
			s := makeServerWithDB(listen, d, bootstrap...)
			s.EncryptionKey = keyBytes
			return s.Start()
		},
	}
	serveCmd.Flags().StringVar(&listen, "listen", ":3000", "listen address")
	serveCmd.Flags().StringSliceVar(&bootstrap, "bootstrap", nil, "bootstrap nodes")
	root.AddCommand(serveCmd)

	storeCmd := &cobra.Command{
		Use:   "store <key> <file>",
		Short: "Store a file locally and broadcast to peers",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, path := args[0], args[1]
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			d, err := dbpkg.Open(dbPath)
			if err != nil {
				return err
			}
			defer d.Close()
			if err := d.Migrate(context.Background()); err != nil {
				return err
			}

			keyBytes, err := loadOrInitKey(d)
			if err != nil {
				return err
			}
			s := makeServerWithDB(listen, d, bootstrap...)
			s.EncryptionKey = keyBytes
			go func() { log.Fatal(s.Start()) }()
			// Wait for connections to establish
			time.Sleep(500 * time.Millisecond)
			if len(bootstrap) > 0 {
				if err := s.waitForPeers(5 * time.Second); err != nil {
					fmt.Printf("Warning: %v. Proceeding with store anyway.\n", err)
				}
			}
			return s.Store(key, f)
		},
	}
	storeCmd.Flags().StringVar(&listen, "listen", ":3000", "listen address")
	storeCmd.Flags().StringSliceVar(&bootstrap, "bootstrap", nil, "bootstrap nodes")
	root.AddCommand(storeCmd)

	getCmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Fetch a file (local or from peers)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			out, _ := cmd.Flags().GetString("out")

			d, err := dbpkg.Open(dbPath)
			if err != nil {
				return err
			}
			defer d.Close()
			if err := d.Migrate(context.Background()); err != nil {
				return err
			}

			keyBytes, err := loadOrInitKey(d)
			if err != nil {
				return err
			}
			s := makeServerWithDB(listen, d, bootstrap...)
			s.EncryptionKey = keyBytes
			go func() { log.Fatal(s.Start()) }()
			// Wait for connections to establish
			time.Sleep(500 * time.Millisecond)
			if len(bootstrap) > 0 {
				if err := s.waitForPeers(5 * time.Second); err != nil {
					fmt.Printf("Warning: %v. Proceeding with get anyway.\n", err)
				}
			}
			_, r, err := s.Get(key)
			if err != nil {
				return err
			}
			var w io.Writer = os.Stdout
			if out != "" {
				of, err := os.Create(out)
				if err != nil {
					return err
				}
				defer of.Close()
				w = of
			}
			_, err = io.Copy(w, r)
			return err
		},
	}
	getCmd.Flags().StringVar(&listen, "listen", ":3000", "listen address")
	getCmd.Flags().StringSliceVar(&bootstrap, "bootstrap", nil, "bootstrap nodes")
	getCmd.Flags().String("out", "", "output file path")
	root.AddCommand(getCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete a file locally and from all peers",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			d, err := dbpkg.Open(dbPath)
			if err != nil {
				return err
			}
			defer d.Close()
			if err := d.Migrate(context.Background()); err != nil {
				return err
			}

			keyBytes, err := loadOrInitKey(d)
			if err != nil {
				return err
			}
			s := makeServerWithDB(listen, d, bootstrap...)
			s.EncryptionKey = keyBytes
			go func() { log.Fatal(s.Start()) }()
			// Wait for connections to establish
			time.Sleep(500 * time.Millisecond)
			if len(bootstrap) > 0 {
				if err := s.waitForPeers(5 * time.Second); err != nil {
					fmt.Printf("Warning: %v. Proceeding with delete anyway.\n", err)
				}
			}
			return s.Delete(key)
		},
	}
	deleteCmd.Flags().StringVar(&listen, "listen", ":3000", "listen address")
	deleteCmd.Flags().StringSliceVar(&bootstrap, "bootstrap", nil, "bootstrap nodes")
	root.AddCommand(deleteCmd)

	filesCmd := &cobra.Command{Use: "files", Short: "File operations"}
	filesListCmd := &cobra.Command{
		Use:   "list",
		Short: "List known files",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := dbpkg.Open(dbPath)
			if err != nil {
				return err
			}
			defer d.Close()
			if err := d.Migrate(context.Background()); err != nil {
				return err
			}
			ff, err := d.ListFiles(context.Background())
			if err != nil {
				return err
			}
			if len(ff) == 0 {
				fmt.Println("No files found.")
				return nil
			}
			fmt.Printf("%-20s\t%-10s\t%s\n", "FILE", "SIZE", "CREATED")
			fmt.Println(strings.Repeat("-", 60))
			for _, f := range ff {
				fmt.Printf("%-20s\t%-10d\t%s\n",
					f.Name,
					f.Size,
					f.CreatedAt.Format("2006-01-02 15:04:05"))
			}
			return nil
		},
	}
	filesCmd.AddCommand(filesListCmd)
	root.AddCommand(filesCmd)

	sharesCmd := &cobra.Command{
		Use:   "shares",
		Short: "List file shares (files stored in other peers)",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := dbpkg.Open(dbPath)
			if err != nil {
				return err
			}
			defer d.Close()
			if err := d.Migrate(context.Background()); err != nil {
				return err
			}
			shares, err := d.ListShares(context.Background())
			if err != nil {
				return err
			}
			if len(shares) == 0 {
				fmt.Println("No shares found.")
				return nil
			}
			fmt.Printf("%-20s\t%-20s\t%-15s\t%-10s\t%s\n", "FILE", "PEER", "DIRECTION", "SIZE", "CREATED")
			fmt.Println(strings.Repeat("-", 100))
			for _, s := range shares {
				fmt.Printf("%-20s\t%-20s\t%-15s\t%-10d\t%s\n",
					s.FileName,
					s.PeerID,
					s.Direction,
					s.FileSize,
					s.CreatedAt.Format("2006-01-02 15:04:05"))
			}
			return nil
		},
	}
	root.AddCommand(sharesCmd)

	// demo: preserves old behavior behind a command
	demoCmd := &cobra.Command{
		Use:   "demo",
		Short: "Run the local 3-node demo",
		RunE: func(cmd *cobra.Command, args []string) error {
			s1 := makeServer(":3000", "")
			s2 := makeServer(":4000", ":3000")
			s3 := makeServer(":5000", ":3000", ":4000")

			go func() { log.Fatal(s1.Start()) }()
			time.Sleep(1 * time.Second)
			go func() { log.Fatal(s2.Start()) }()
			time.Sleep(1 * time.Second)
			go s3.Start()
			time.Sleep(1 * time.Second)

			key := "coolpicture.jpg"
			data := bytes.NewReader([]byte("my big data file here!"))
			_ = s3.Store(key, data)
			_ = s3.Delete(key)
			_, r, err := s3.Get(key)
			if err != nil {
				return err
			}
			b, err := io.ReadAll(r)
			if err != nil {
				return err
			}
			fmt.Println(string(b))
			return nil
		},
	}
	root.AddCommand(demoCmd)

	return root
}

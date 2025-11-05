package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
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
			time.Sleep(500 * time.Millisecond)
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
			time.Sleep(500 * time.Millisecond)
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
		Short: "Delete a file locally",
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
			time.Sleep(500 * time.Millisecond)
			return s.store.Delete(key)
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
			for _, f := range ff {
				fmt.Printf("%s\t%s\t%d\t%s\n", f.ID, f.Name, f.Size, f.LocalPath)
			}
			return nil
		},
	}
	filesCmd.AddCommand(filesListCmd)
	root.AddCommand(filesCmd)

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
			_ = s3.store.Delete(key)
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

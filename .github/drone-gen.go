package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
)

type Driver struct {
	Name       string
	InitSource string
	TestSource string
}

type Database struct {
	ImageName     string
	ImageVersions []string
	Drivers       []Driver
}

func main() {
	log.SetFlags(0)
	flag.Parse()

	goImages := []string{
		"golang:1.6",
		"golang:1.7",
		"captncraig/go-tip",
	}

	databases := []Database{
		// https://www.postgresql.org/support/versioning/
		{
			"postgres",
			[]string{
				"9.2",
				"9.3",
				"9.4",
				"9.5",
				"9.6",
			}, []Driver{
				{
					"postgres",
					"postgres://reform-user:reform-password123@127.0.0.1/reform-database?sslmode=disable&TimeZone=UTC",
					"postgres://reform-user:reform-password123@127.0.0.1/reform-database?sslmode=disable&TimeZone=America/New_York",
				},
			},
		},

		// https://www.mysql.com/support/supportedplatforms/database.html
		{
			"mysql",
			[]string{
				"5.5",
				"5.6",
				"5.7",
				"8.0",
			}, []Driver{
				// ANSI mode
				{
					"mysql",
					"root@/reform-database?parseTime=true&time_zone='UTC'&sql_mode='ANSI'&multiStatements=true",
					"reform-user:reform-password123@/reform-database?parseTime=true&time_zone='America%2FNew_York'&sql_mode='ANSI'",
				},

				// TRADITIONAL mode + interpolateParams=true
				{
					"mysql",
					"root@/reform-database?parseTime=true&time_zone='UTC'&sql_mode='ANSI'&multiStatements=true",
					"reform-user:reform-password123@/reform-database?parseTime=true&time_zone='America%2FNew_York'&sql_mode='TRADITIONAL'&interpolateParams=true",
				},
			},
		},

		{
			"sqlite3",
			[]string{
				"dummy",
			},
			[]Driver{
				{
					"sqlite3",
					"/tmp/reform-test.sqlite3",
					"/tmp/reform-test.sqlite3",
				},
			},
		},

		{
			"mssql",
			[]string{
				"latest",
			},
			[]Driver{
				{
					"mssql",
					"server=localhost;user id=sa;password=reform-password123;database=reform-database",
					"server=localhost;user id=sa;password=reform-password123;database=reform-database",
				},
			},
		},
	}

	const start = `# generated with 'go run .github/drone-gen.go'`

	var buf bytes.Buffer
	buf.WriteString(start + "\n")
	buf.WriteString(fmt.Sprintf("# Go %s\n", goImages))
	for _, db := range databases {
		drivers := make([]string, 0, len(db.Drivers))
		for _, d := range db.Drivers {
			drivers = append(drivers, d.Name)
		}
		buf.WriteString(fmt.Sprintf("# %s %s %s\n", db.ImageName, db.ImageVersions, drivers))
	}

	buf.WriteString("matrix:\n")
	buf.WriteString("  include:\n")

	var count int
	for _, g := range goImages {
		for _, db := range databases {
			for _, v := range db.ImageVersions {
				for _, d := range db.Drivers {
					fmt.Fprintf(&buf, "    - {\n")
					fmt.Fprintf(&buf, "        GO: %q, DATABASE: %s, VERSION: %s, REFORM_DRIVER: %s,\n",
						g, db.ImageName, v, d.Name)
					fmt.Fprintf(&buf, "        REFORM_INIT_SOURCE: %q,\n", d.InitSource)
					fmt.Fprintf(&buf, "        REFORM_TEST_SOURCE: %q", d.TestSource)
					fmt.Fprintf(&buf, "\n      }\n")
					count++
				}
			}
		}
	}

	// update file
	const filename = ".drone.yml"
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	i := bytes.Index(b, []byte(start))
	if i < 0 {
		log.Fatalf("failed to find %q in %s", start, filename)
	}
	b = append(b[:i], buf.Bytes()...)
	err = ioutil.WriteFile(filename, b, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s matrix updated with %d combinations.", filename, count)
}

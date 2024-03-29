package docker

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"sync"
	"time"
)

func Postgres(ctx context.Context, wg *sync.WaitGroup, dbname string) (conStr string, err error) {
	conStr, _, _, err = PostgresWithNetwork(ctx, wg, dbname)
	return
}

func PostgresWithNetwork(ctx context.Context, wg *sync.WaitGroup, dbname string) (conStr string, ip string, port string, err error) {
	log.Println("start postgres")
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "postgres:11.2",
			Env: map[string]string{
				"POSTGRES_DB":       dbname,
				"POSTGRES_PASSWORD": "pw",
				"POSTGRES_USER":     "usr",
			},
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor: wait.ForAll(
				wait.ForListeningPort("5432/tcp"),
				wait.ForLog("database system is ready to accept connections"),
			),
			Tmpfs: map[string]string{"/var/lib/postgresql/data": "rw"},
			//SkipReaper: true,
		},
		Started: true,
	})
	if err != nil {
		return "", "", "", err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		log.Println("DEBUG: remove container postgres", c.Terminate(context.Background()))
	}()

	ip, err = c.ContainerIP(ctx)
	if err != nil {
		return "", "", "", err
	}
	temp, err := c.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return "", "", "", err
	}
	port = temp.Port()
	conStr = fmt.Sprintf("postgres://usr:pw@%s:%s/%s?sslmode=disable", ip, "5432", dbname)

	err = retry(1*time.Minute, func() error {
		log.Println("try pg conn", conStr)
		db, err := sql.Open("postgres", conStr)
		if err != nil {
			return err
		}
		err = db.Ping()
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Println("ERROR:", err)
		return "", "", "", err
	}

	return conStr, ip, port, err
}

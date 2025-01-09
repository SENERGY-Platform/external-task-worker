package test

import (
	"context"
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/camunda"
	"github.com/SENERGY-Platform/external-task-worker/lib/camunda/shards"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/lib/prometheus"
	"github.com/SENERGY-Platform/external-task-worker/lib/test/docker"
	"github.com/SENERGY-Platform/external-task-worker/lib/test/mock"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestFetch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	defer wg.Wait()
	defer cancel()

	pgConn, err := docker.Postgres(ctx, &wg, "test")
	if err != nil {
		t.Error(err)
		return
	}

	s, err := shards.New(pgConn, nil)
	if err != nil {
		t.Error(err)
		return
	}

	_, camundaPgIp, _, err := docker.PostgresWithNetwork(ctx, &wg, "camunda")
	if err != nil {
		t.Error(err)
		return
	}

	camundaUrl, err := docker.Camunda(ctx, &wg, camundaPgIp, "5432")
	if err != nil {
		t.Error(err)
		return
	}
	err = s.EnsureShard(camundaUrl)
	if err != nil {
		t.Error(err)
		return
	}
	shard, err := s.EnsureShardForUser("owner")
	if err != nil {
		t.Error(err)
		return
	}

	c := camunda.NewCamundaWithShards(util.Config{
		CamundaFetchLockDuration: 60000,
		CamundaTopic:             "pessimistic",
		CamundaWorkerTasks:       10,
	}, mock.Kafka, prometheus.NewMetrics("test", nil), s)

	providerCtx, _ := context.WithTimeout(ctx, 1*time.Minute)
	taskChan, errChan, err := c.ProvideTasks(providerCtx)
	if err != nil {
		t.Error(err)
		return
	}
	go func() {
		time.Sleep(25 * time.Second)
		var pid string
		testCreateProcess(camundaUrl, &pid)(t)
		testStartDeployment(shard, pid)(t)
		log.Println(time.Now(), "process started")
	}()
	go func() {
		for err := range errChan {
			if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
				t.Error(err)
			}
		}
	}()
	receivedTaskBatchCounts := []int{}
	for task := range taskChan {
		receivedTaskBatchCounts = append(receivedTaskBatchCounts, len(task))
		log.Println(time.Now(), len(task))
	}
	//2 long polls without task (10s + 10s)
	//after 25s (5s after last long poll) 1 task found
	//3 long polls without task (10s + 10s + 10s)
	//stop after 60s, 5s before next empty poll
	if !reflect.DeepEqual(receivedTaskBatchCounts, []int{0, 0, 1, 0, 0, 0}) {
		t.Error("unexpected batch count or task count in batches")
	}
}

func TestGetTask(t *testing.T) {
	temp := util.GetId
	util.GetId = func() string {
		return "test-worker"
	}
	defer func() {
		util.GetId = temp
	}()

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	defer wg.Wait()
	defer cancel()

	pgConn, err := docker.Postgres(ctx, &wg, "test")
	if err != nil {
		t.Error(err)
		return
	}

	s, err := shards.New(pgConn, nil)
	if err != nil {
		t.Error(err)
		return
	}

	_, camundaPgIp, _, err := docker.PostgresWithNetwork(ctx, &wg, "camunda")
	if err != nil {
		t.Error(err)
		return
	}

	camundaUrl, err := docker.Camunda(ctx, &wg, camundaPgIp, "5432")
	if err != nil {
		t.Error(err)
		return
	}
	err = s.EnsureShard(camundaUrl)
	if err != nil {
		t.Error(err)
		return
	}
	shard, err := s.EnsureShardForUser("owner")
	if err != nil {
		t.Error(err)
		return
	}

	_, err = mock.Kafka.NewProducer(ctx, util.Config{})
	if err != nil {
		t.Fatal(err)
	}

	var pid string
	var tasks []messages.CamundaExternalTask
	t.Run("create normal process", testCreateProcess(camundaUrl, &pid))
	t.Run("start process", testStartDeployment(shard, pid))
	t.Run("run get task", testGetTasks(s, &tasks))
	t.Run("complete task", testCompleteTask(s, tasks))
	t.Run("check incidents", testCheckIncidents())
}

func testCheckIncidents() func(t *testing.T) {
	return func(t *testing.T) {
		if incidents := mock.Kafka.GetProduced("incidents"); len(incidents) > 0 {
			t.Fatal(incidents)
		}
	}
}

func testCompleteTask(s *shards.Shards, tasks []messages.CamundaExternalTask) func(t *testing.T) {
	return func(t *testing.T) {
		if len(tasks) < 1 {
			t.Fatal(tasks)
		}
		err := camunda.NewCamundaWithShards(util.Config{
			KafkaIncidentTopic: "incidents",
		}, mock.Kafka, prometheus.NewMetrics("test", nil), s).CompleteTask(messages.TaskInfo{
			WorkerId:            "test-worker",
			TaskId:              tasks[0].Id,
			ProcessInstanceId:   tasks[0].ProcessInstanceId,
			ProcessDefinitionId: tasks[0].ProcessDefinitionId,
			TenantId:            tasks[0].TenantId,
		}, "result", map[string]interface{}{"r": 255, "g": 0, "b": 100})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func testGetTasks(s *shards.Shards, tasks *[]messages.CamundaExternalTask) func(t *testing.T) {
	return func(t *testing.T) {
		var err error
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		taskChan, errChan, err := camunda.NewCamundaWithShards(util.Config{
			CamundaFetchLockDuration: 60000,
			CamundaTopic:             "pessimistic",
			CamundaWorkerTasks:       10,
		}, mock.Kafka, prometheus.NewMetrics("test", nil), s).ProvideTasks(ctx)
		if err != nil {
			t.Fatal(err)
		}
		for err := range errChan {
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Error(err)
			}
		}
		for batch := range taskChan {
			if len(batch) > 0 {
				*tasks = batch
			}
		}
		if len(*tasks) != 1 {
			t.Fatal(*tasks)
		}
	}
}

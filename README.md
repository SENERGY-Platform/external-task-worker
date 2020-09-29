consumes external tasks from camunda bpmn execution engine and executes them in the sepl-platform 
by publishing a message on kafka and relaying the response back to camunda

formerly known as camundaworker

uses a database to find camunda shards
ref https://github.com/SENERGY-Platform/camunda-engine-wrapper for details about the initialization of the sharding-db
package bgtasker

import "time"

const (
	RunnerTypeBlock = "block"
	RunnerTypeDrop  = "drop"
)

const (
	defaultAsyncWorkersNum    = 8
	defaultAsyncBufferSize    = 1024
	defaultAsyncWorkerTimeout = 2 * time.Second

	defaultTaskType = RunnerTypeDrop
)

type TaskConfig struct {
	RunnerType         string        `yaml:"runner_type"`
	AsyncWorkersNum    int           `yaml:"async_workers_num"`
	AsyncBufferSize    int           `yaml:"async_buffer_size"`
	AsyncWorkerTimeout time.Duration `yaml:"async_worker_timeout"`
}

func (c *TaskConfig) Normalize() {
	if c.AsyncWorkersNum == 0 {
		c.AsyncWorkersNum = defaultAsyncWorkersNum
	}

	if c.AsyncBufferSize == 0 {
		c.AsyncBufferSize = defaultAsyncBufferSize
	}

	if c.AsyncWorkerTimeout == 0 {
		c.AsyncWorkerTimeout = defaultAsyncWorkerTimeout
	}
	if c.RunnerType == "" {
		c.RunnerType = defaultTaskType
	}
}

type Config struct {
	Tasks map[string]*TaskConfig `yaml:"tasks"`
}

func DefaultTaskConfig() *TaskConfig {
	return &TaskConfig{
		RunnerType:         defaultTaskType,
		AsyncWorkersNum:    defaultAsyncWorkersNum,
		AsyncBufferSize:    defaultAsyncBufferSize,
		AsyncWorkerTimeout: defaultAsyncWorkerTimeout,
	}
}

func NewConfig() *Config {
	return &Config{
		Tasks: make(map[string]*TaskConfig),
	}
}

func (c *Config) Prepare() *Config {
	for name, taskConfig := range c.Tasks {
		if taskConfig == nil {
			c.Tasks[name] = DefaultTaskConfig()

			continue
		}

		taskConfig.Normalize()

		c.Tasks[name] = taskConfig
	}

	return c
}

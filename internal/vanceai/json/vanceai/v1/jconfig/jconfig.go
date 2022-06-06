package jconfig

type JConfig interface {
	Job() string
}

type SingleJob struct {
	Job_    string `json:"job"`
	Config_ Config `json:"config"`
}

func (j *SingleJob) Job() string {
	return j.Job_
}

func NewSingleJob(f Feature) *SingleJob {
	return &SingleJob{
		Job_:    f.Name(),
		Config_: f.Config(),
	}
}

type Workflow struct {
	Job_     string    `json:"job"`
	Features []Feature `json:"config"`
}

func NewWorkflow(f []Feature) *Workflow {
	if f == nil {
		f = make([]Feature, 0)
	}

	return &Workflow{
		Job_:     "workflow",
		Features: f,
	}
}

func (w *Workflow) Job() string { //TODO: test for check if encodes Job
	return w.Job_
}

type Feature interface {
	Name() string
	Config() Config
}

type ToongineerCartoonizer struct {
	Name_   string                       `json:"job"`
	Config_ *ToongineerCartoonizerConfig `json:"config"`
}

func NewToongineerCartoonizer() *ToongineerCartoonizer {
	return &ToongineerCartoonizer{
		Name_:   "cartoonize",
		Config_: newToongineerCartoonizerConfig(),
	}
}

func (tc *ToongineerCartoonizer) Name() string {
	return tc.Name_
}

func (tc *ToongineerCartoonizer) Config() Config {
	return tc.Config_
}

type Config interface {
	Module() string
	ModuleParams() ModuleParams
	SetOutParams(p *OutParams)
}

type ToongineerCartoonizerConfig struct {
	Module_       string               `json:"module"`
	ModuleParams_ *ModuleParamsDefault `json:"module_params"`
	OutParams_    *OutParams           `json:"out_params,omitempty"`
}

func newToongineerCartoonizerConfig() *ToongineerCartoonizerConfig {
	return &ToongineerCartoonizerConfig{
		Module_:       "cartoonize",
		ModuleParams_: newModuleParamsDefault("CartoonizeStable"),
		OutParams_:    nil,
	}
}

func (c *ToongineerCartoonizerConfig) Module() string {
	return c.Module_
}

func (c *ToongineerCartoonizerConfig) ModuleParams() ModuleParams {
	return c.ModuleParams_
}

func (c *ToongineerCartoonizerConfig) SetOutParams(p *OutParams) {
	c.OutParams_ = p
}

type ModuleParams interface {
	ModelName() string
}

type ModuleParamsDefault struct {
	ModelName_ string `json:"model_name"`
}

func newModuleParamsDefault(modelName string) *ModuleParamsDefault {
	return &ModuleParamsDefault{ModelName_: modelName}
}

func (p *ModuleParamsDefault) ModelName() string {
	return p.ModelName_
}

type OutParams struct {
	SomeField string `json:"some_field"`
}

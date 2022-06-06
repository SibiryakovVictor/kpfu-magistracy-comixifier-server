package image

import "comixifier/internal/vanceai/json/vanceai/v1/jconfig"

type Processor interface {
	Map() jconfig.Feature
}

type Cartoonizer struct {
}

func NewCartoonizer() *Cartoonizer {
	return &Cartoonizer{}
}

func (c *Cartoonizer) Map() jconfig.Feature {
	return jconfig.NewToongineerCartoonizer()
}

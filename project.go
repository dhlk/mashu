package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type PlanClip struct {
	Region Region
	SrcKey *string
	Source *Source
}

type PlanStack []string

type PlanBlend struct {
	Name        string
	Attachments Attachments
}

type PlanConcat []string

type Plan struct {
	Name   string
	Blend  *PlanBlend
	Clip   *PlanClip
	Concat *PlanConcat
	Stack  *PlanStack
}

type Project struct {
	Path    string
	Catalog Catalog
	Format  Format
	Plan    Plan
}

// TODO global stamping toggle? maybe disalbe when format has no stamp

func NewProject(path string, c Catalog) (p Project, err error) {
	var fi os.FileInfo
	if fi, err = os.Stat(path); err != nil || !fi.IsDir() {
		if err != nil {
			err = fmt.Errorf("mashu.NewProject: unable to stat project directory ('%s'): %w", path, err)
			return
		}
		err = fmt.Errorf("mashu.NewProject: project must be a directory ('%s')", path)
		return
	}

	p.Path = path
	p.Catalog = c

	if err = decodeJsonFromFile(filepath.Join(p.Path, "format.json"), &p.Format); err != nil {
		err = fmt.Errorf("mashu.NewProject: unable to load format ('%s/format.json'): %w", path, err)
		return
	}
	if err = p.Format.Valid(); err != nil {
		err = fmt.Errorf("mashu.NewProject: invalid format ('%s/format.json'): %w", path, err)
		return
	}

	if err = decodeJsonFromFile(filepath.Join(p.Path, "plan.json"), &p.Plan); err != nil {
		err = fmt.Errorf("mashu.NewProject: unable to load plan ('%s/plan.json'): %w", path, err)
		return
	}

	return
}

func (p Project) Execute() error {
	return p.executePlan(p.Plan)
}

func (p Project) executePlanByName(name string) error {
	path := filepath.Join(p.Path, "plan", fmt.Sprintf("%s.json", name))

	var plan Plan
	if err := decodeJsonFromFile(path, &plan); err != nil {
		return fmt.Errorf("mashu.NewProject: unable to load plan ('%s'): %w", path, err)
	}

	return p.executePlan(plan)
}

func (p Project) getInput(name string) (i Input, err error) {
	i = Input(filepath.Join(p.Path, "render",
		fmt.Sprintf("%s.%s", name, strings.ToLower(p.Format.Format))))
	err = i.Valid()
	return
}

func (p Project) getOutput(name string) (o Output, err error) {
	o = Output(filepath.Join(p.Path, "render",
		fmt.Sprintf("%s.%s", name, strings.ToLower(p.Format.Format))))
	err = o.Valid()
	return
}

func (p Project) executePlan(plan Plan) (err error) {
	var o Output
	if o, err = p.getOutput(plan.Name); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return nil
		}
		return
	}

	if plan.Clip != nil {
		return p.executePlanClip(*plan.Clip, o)
	}
	if plan.Blend != nil {
		return p.executePlanBlend(*plan.Blend, o)
	}
	if plan.Concat != nil {
		return p.executePlanConcat(*plan.Concat, o)
	}
	if plan.Stack != nil {
		return p.executePlanStack(*plan.Stack, o)
	}

	return fmt.Errorf("mashu.Project.executePlan: invalid plan ('%s')", plan.Name)
}

func (p Project) executePlanClip(clip PlanClip, output Output) (err error) {
	if err = clip.Region.Valid(); err != nil {
		return
	}

	var s Source
	if clip.Source != nil {
		s = *clip.Source
	} else if clip.SrcKey != nil {
		if s, err = p.Catalog.Lookup(*clip.SrcKey); err != nil {
			return
		}
	} else {
		return fmt.Errorf("mashu.Project.executePlanClip: ivalid clip plan (no source)")
	}

	if err = s.Valid(); err != nil {
		return
	}

	if err = renderClip(context.TODO(), p.Format, s, clip.Region, output); err != nil {
		return
	}

	return
}

func (p Project) executePlanConcat(concat PlanConcat, output Output) (err error) {
	inputs := make([]Input, len(concat))
	for i, name := range concat {
		if err = p.executePlanByName(name); err != nil {
			return
		}
		if inputs[i], err = p.getInput(name); err != nil {
			return
		}
	}

	if err = renderConcat(context.TODO(), output, inputs); err != nil {
		return
	}

	return
}

func (p Project) executePlanStack(stack PlanStack, output Output) (err error) {
	inputs := make([]Input, len(stack))
	for i, name := range stack {
		if err = p.executePlanByName(name); err != nil {
			return
		}
		if inputs[i], err = p.getInput(name); err != nil {
			return
		}
	}

	if err = renderStack(context.TODO(), p.Format, output, inputs); err != nil {
		return
	}

	return
}

func (p Project) executePlanBlend(blend PlanBlend, output Output) (err error) {
	attachments := make(Attachments)
	for k, name := range blend.Attachments {
		// TODO this is a mess; blender stuff needs rewrite
		if err = p.executePlanByName(string(name)); err != nil {
			return
		}
		if attachments[k], err = p.getInput(string(name)); err != nil {
			return
		}
	}

	if err = renderBlend(context.TODO(), blend.Name, p.Format, output, attachments); err != nil {
		return
	}

	return
}

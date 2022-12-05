package main

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type PlanSegment struct {
	Clip  *Region `json:",omitempty"`
	Stack *struct {
		Count    uint
		Duration Region
	} `json:",omitempty"`
	Blend   *string `json:",omitempty"`
	Tickets uint
}

type PlanGeneratorParameters struct {
	Target         Duration
	Alignment      Duration
	MaxConcat      uint
	RequiredTags   []string
	DisallowedTags []string
	Segments       []PlanSegment
}

func (g PlanGeneratorParameters) PullSegmentFunc() (func() PlanSegment, error) {
	type border struct {
		Border  uint
		Segment PlanSegment
	}

	if len(g.Segments) == 0 {
		return nil, fmt.Errorf("mashu.PlanGeneratorParameters.PullSegmentFunc: must specify segments")
	}

	tickets := uint(0)
	borders := make([]border, len(g.Segments))

	for i, s := range g.Segments {
		tickets += s.Tickets
		borders[i] = border{tickets, s}
	}

	return func() PlanSegment {
		n := uint(rand.Intn(int(tickets))) + 1
		for _, b := range borders {
			if n <= b.Border {
				return b.Segment
			}
		}

		panic(nil)
	}, nil
}

func (g PlanGeneratorParameters) PullSourceFunc(p Project) (fn func(n int) ([]Source, error), err error) {
	var keys []string
	if keys, err = p.Catalog.Keys(context.TODO()); err != nil {
		return
	}

	if len(g.RequiredTags)+len(g.DisallowedTags) > 0 {
		var validKeys []string
		for _, key := range keys {
			var s Source
			if s, err = p.Catalog.Lookup(key); err != nil {
				return
			}

			if len(validRegions(g, s.Regions)) > 0 {
				validKeys = append(validKeys, key)
			}
		}
		keys = validKeys
	}

	if len(keys) == 0 {
		err = fmt.Errorf("mashu.PlanGeneratorParameters.PullSourceFunc: no keys match generator restrictions")
		return
	}

	idx := 0
	fn = func(n int) (sources []Source, err error) {
		sources = make([]Source, n)
		for i := 0; i < n; i++ {
			if idx == len(keys) {
				idx = 0
				rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
			}
			if sources[i], err = p.Catalog.Lookup(keys[idx]); err != nil {
				return
			}
			idx += 1
		}

		return
	}
	return
}

func planCreate(p Project, prefix string) (name string, f *os.File, err error) {
	for {
		b := make([]byte, 16)
		if _, err = crand.Read(b); err != nil {
			return
		}
		name = fmt.Sprintf("%s-%x", prefix, b)
		path := filepath.Join(p.Path, "plan", name+".json")
		if _, err = os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			err = nil
			continue
		}

		f, err = os.Create(path)
		return
	}
}

func validRegions(g PlanGeneratorParameters, regions []TaggedRegion) (validRegions []Region) {
region:
	for _, r := range regions {
	requiredTag:
		for _, requiredTag := range g.RequiredTags {
			for _, includedTag := range r.Tags {
				if requiredTag == includedTag {
					continue requiredTag
				}
			}
			continue region
		}
		for _, disallowedTag := range g.DisallowedTags {
			for _, includedTag := range r.Tags {
				if disallowedTag == includedTag {
					continue region
				}
			}
		}
		validRegions = append(validRegions, r.Region)
	}

	return
}

func planConcat(p Project, names []string) (name string, err error) {
	plan := Plan{Concat: &PlanConcat{Input: names}}

	var f *os.File
	if plan.Name, f, err = planCreate(p, "concat"); err != nil {
		return
	}
	defer f.Close()

	name = plan.Name

	e := json.NewEncoder(f)
	if err = e.Encode(plan); err != nil {
		return
	}

	return
}

func planStack(p Project, target Duration, names []string) (name string, err error) {
	plan := Plan{Stack: &PlanStack{Input: names, Duration: target}}

	var f *os.File
	if plan.Name, f, err = planCreate(p, "stack"); err != nil {
		return
	}
	defer f.Close()

	name = plan.Name

	e := json.NewEncoder(f)
	if err = e.Encode(plan); err != nil {
		return
	}

	return
}

func planBlend(p Project, g PlanGeneratorParameters, blend string, pullSource func(n int) ([]Source, error)) (name string, d Duration, err error) {
	att := make(Attachments)
	switch blend {
	case "demon-slayer-scrolls":
		inDuration := [16]Duration{
			Duration{time.Second * 7},
			Duration{time.Second * 16},
			Duration{time.Second * 16},
			Duration{time.Second * 16},
			Duration{time.Second * 16},
			Duration{time.Second * 16},
			Duration{time.Second * 16},
			Duration{time.Second * 16},
			Duration{time.Second * 16},
			Duration{time.Second * 16},
			Duration{time.Second * 16},
			Duration{time.Second * 16},
			Duration{time.Second * 16},
			Duration{time.Second * 16},
			Duration{time.Second * 16},
			Duration{time.Second * 16},
		}
		outDuration := Duration{time.Second * 20}
		d = outDuration

		var s []Source
		if s, err = pullSource(16); err != nil {
			return
		}

		for i, src := range s {
			var clipName string
			if clipName, _, err = planClip(p, g, inDuration[i], src); err != nil {
				return
			}
			att[fmt.Sprintf("video.%03d", i+1)] = Input(clipName)
		}
	default:
		err = fmt.Errorf("mashu.planBlend: invalid blend in generator configuration")
		return
	}

	plan := Plan{Blend: &PlanBlend{Name: blend, Attachments: att}}

	var f *os.File
	if plan.Name, f, err = planCreate(p, "blend"); err != nil {
		return
	}
	defer f.Close()

	name = plan.Name

	e := json.NewEncoder(f)
	if err = e.Encode(plan); err != nil {
		return
	}

	return

	// TODO this is super hardcoded; embed json to configure this?
	//func planBlend(p Project, blend string, attachments map[string]string) (name string, d Duration, err error) {
	return
}

func planClip(p Project, g PlanGeneratorParameters, target Duration, s Source) (name string, d Duration, err error) {
	plan := Plan{Clip: &PlanClip{SrcKey: &s.Key}}

	var f *os.File
	if plan.Name, f, err = planCreate(p, "clip"); err != nil {
		return
	}
	defer f.Close()

	name = plan.Name
	d = target

	regions := validRegions(g, s.Regions)
	region := regions[rand.Intn(len(regions))]
	regionDuration := region.Duration()

	if regionDuration > d.Duration {
		plan.Clip.Region.Start = region.Start.Add(Duration{time.Duration(
			rand.Int63n(int64(region.End.Duration - region.Start.Duration - d.Duration)))})
		plan.Clip.Region.End = plan.Clip.Region.Start.Add(d)
	} else {
		d = Duration{regionDuration.Truncate(g.Alignment.Duration)}
		plan.Clip.Region.Start = region.Start
		plan.Clip.Region.End = region.Start.Add(d)
	}

	e := json.NewEncoder(f)
	if err = e.Encode(plan); err != nil {
		return
	}

	return
}

func (p Project) Generate() (err error) {
	rand.Seed(time.Now().UnixNano())
	var g PlanGeneratorParameters
	if err = decodeJsonFromFile(filepath.Join(p.Path, "generator.json"), &g); err != nil {
		return
	}

	var pullSeg func() PlanSegment
	if pullSeg, err = g.PullSegmentFunc(); err != nil {
		return
	}

	var pullSource func(n int) ([]Source, error)
	if pullSource, err = g.PullSourceFunc(p); err != nil {
		return
	}

	var d Duration
	var names [][]string
	names = append(names, make([]string, 0))
	for d.Duration < g.Target.Duration {
		for layer, group := range names {
			if len(group) == int(g.MaxConcat) {
				var name string
				if name, err = planConcat(p, group); err != nil {
					return
				}

				if len(names) == layer+1 {
					names = append(names, []string{name})
				} else {
					names[layer+1] = append(names[layer+1], name)
				}

				names[layer] = make([]string, 0)
			}
		}

		t := pullSeg()
		if t.Clip != nil {
			var s []Source
			if s, err = pullSource(1); err != nil {
				return
			}

			var name string
			var clipDuration Duration
			if name, clipDuration, err = planClip(p, g, t.Clip.RandomTruncated(g.Alignment), s[0]); err != nil {
				return
			}
			d = d.Add(clipDuration)
			names[0] = append(names[0], name)
		} else if t.Stack != nil {
			var s []Source
			if s, err = pullSource(int(t.Stack.Count)); err != nil {
				return
			}

			target := t.Stack.Duration.RandomTruncated(g.Alignment)
			snames := make([]string, t.Stack.Count)
			for i, source := range s {
				if snames[i], _, err = planClip(p, g, target, source); err != nil {
					return
				}
			}

			var name string
			if name, err = planStack(p, target, snames); err != nil {
				return
			}
			d = d.Add(target)
			names[0] = append(names[0], name)
		} else if t.Blend != nil {
			var name string
			var target Duration
			if name, target, err = planBlend(p, g, *t.Blend, pullSource); err != nil {
				return
			}

			d = d.Add(target)
			names[0] = append(names[0], name)
		} else {
			return fmt.Errorf("mashu.Project.Generate: invalid segment in generator configuration")
		}
	}

	for len(names) > 1 {
		if len(names[0]) == int(g.MaxConcat) {
			var name string
			if name, err = planConcat(p, names[0]); err != nil {
				return
			}
			names[1] = append(names[1], name)
			names = names[1:]
		} else if len(names[0])+len(names[1]) <= int(g.MaxConcat) {
			names[1] = append(names[1], names[0]...)
			names = names[1:]
		} else {
			split := len(names[1]) - (int(g.MaxConcat) - len(names[0]))
			names[0] = append(names[1][split:], names[0]...)
			names[1] = names[1][:split]
		}
	}

	var rootPlan string
	if len(names[0]) == 1 {
		rootPlan = names[0][0]
	} else if len(names[0]) > 1 {
		if rootPlan, err = planConcat(p, names[0]); err != nil {
			return
		}
	} else {
		return fmt.Errorf("mashu.Project.Generate: unable to determine root plan")
		return
	}

	if err = os.Symlink(filepath.Join("plan", rootPlan+".json"),
		filepath.Join(p.Path, "plan.json")); err != nil {
		return
	}
	if err = os.Symlink(filepath.Join("render", rootPlan+"."+strings.ToLower(p.Format.Format)),
		filepath.Join(p.Path, "output."+strings.ToLower(p.Format.Format))); err != nil {
		return
	}

	return
}

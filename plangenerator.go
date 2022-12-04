package main

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
)

type PlanSegment struct {
	Operation     string
	DurationRange *Region
	Options       any
	Tickets       uint
}

type PlanGeneratorParameters struct {
	Target         Duration
	Alignment      Duration
	MaxConcat      uint
	RequiredTags   []string
	DisallowedTags []string
	Segments       []PlanSegment
}

func (g PlanGeneratorParameters) PullFunc() (func() PlanSegment, error) {
	type border struct {
		Border  uint
		Segment PlanSegment
	}

	if len(g.Segments) == 0 {
		return nil, fmt.Errorf("mashu.PlanGeneratorParameters.PullFunc: must specify segments")
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
			if b.Border <= n {
				return b.Segment
			}
		}

		panic(nil)
	}, nil
}

func (p Project) Generate() (err error) {
	var g PlanGeneratorParameters
	if err = decodeJsonFromFile(filepath.Join(p.Path, "generator.json"), &g); err != nil {
		return
	}

	var pull func() PlanSegment
	if pull, err = g.PullFunc(); err != nil {
		return
	}

	var keys []string
	if keys, err = p.Catalog.Keys(context.TODO()); err != nil {
		return
	}

	// TODO
	if false {
		keys[0] = ""
		pull()
	}
	return
}

// TODO tag filter input needs to be somewhere for the generator config
/*
/var/mashu/
	catalog/
		keys
		{fsmap key}/source.json
		{fsmap key}/override.{ass,srt} (when an override is needed for this entity)
	project/{project}/
		generator.json (generator params; rename?)
*/
/*
{
	"MaxSequenceLength": 8,
	"ClipAlignment": '0.1s', // duration precision
	dist: [
		{"name": "type 1", "tickets": 1},
		{"name": "type 2", "tickets": 50},
		{"name": "type 3", "tickets": 50},
	],
	def: {
		"type 1": {
			operation: "blend",
			options: {
				"name": "demonslayer-scroll"
			}
		},
		"type 2": {
			operation: "stack",
			"duration": {"from": "0.2s", "to": "0.5s"},
			options: {
				"concurrent": [
					{"count": 4, "tickets": 9},
					{"count": 9, "tickets": 1},
				],
			}
		},
		"type 3": {
			operation: "clip",
			"duration": {"from": "0.2s", "to": "0.5s"},
		},
	}
}
*/

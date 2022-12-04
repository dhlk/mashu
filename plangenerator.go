package main

type PlanSegment struct {
	Operation     string
	DurationRange *Region
	Options       any
	Tickets       uint
}

type PlanGeneratorParameters struct {
	Target            Duration
	Alignment         Duration
	MaxSequenceLength uint
	Segments          []PlanSegment
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

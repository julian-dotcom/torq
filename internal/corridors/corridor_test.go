package corridors

import (
	"reflect"
	"testing"
)

func Test_GetBestCorridorFlag(t *testing.T) {
	// README: Assume tag `test` has `tagId: 5`
	testTag := 5
	testNode1 := 1
	testNode2 := 2
	testChannel1 := 1
	tagActive := 1
	tagInactive := 0
	corridorId := 1
	corridorStagingCache := make(map[int]map[CorridorKey]Corridor, 0)
	testTagForPeer1Corridor := Corridor{
		CorridorTypeId: Tag().CorridorTypeId,
		CorridorId:     corridorId,
		Flag:           tagActive,
		ReferenceId:    &testTag,
		FromNodeId:     &testNode1,
	}
	testTagForPeer1Corridor.Priority = calculatePriority(testTagForPeer1Corridor)
	addToCorridorCache(testTagForPeer1Corridor, &corridorStagingCache)
	corridorId = corridorId + 1
	testTagForPeer1Channel1Corridor := Corridor{
		CorridorTypeId: Tag().CorridorTypeId,
		CorridorId:     corridorId,
		Flag:           tagInactive,
		ReferenceId:    &testTag,
		FromNodeId:     &testNode1,
		ChannelId:      &testChannel1,
	}
	testTagForPeer1Channel1Corridor.Priority = calculatePriority(testTagForPeer1Channel1Corridor)
	addToCorridorCache(testTagForPeer1Channel1Corridor, &corridorStagingCache)
	corridorId = corridorId + 1
	testTagForPeer2Corridor := Corridor{
		CorridorTypeId: Tag().CorridorTypeId,
		CorridorId:     corridorId,
		Flag:           tagActive,
		ReferenceId:    &testTag,
		FromNodeId:     &testNode2,
	}
	testTagForPeer2Corridor.Priority = calculatePriority(testTagForPeer2Corridor)
	addToCorridorCache(testTagForPeer2Corridor, &corridorStagingCache)
	finalizeCorridorCacheByType(Tag(), &corridorStagingCache)

	tests := []struct {
		name    string
		input   CorridorKey
		want    int
		wantErr bool
	}{{
		name: "Find corridor with FromNodeId 1",
		input: CorridorKey{
			CorridorType: Tag(),
			ReferenceId:  5,
			FromNodeId:   1,
		},
		want: 1,
	}, {
		name: "Find corridor with FromNodeId 1 and ChannelId 1",
		input: CorridorKey{
			CorridorType: Tag(),
			FromNodeId:   1,
			ReferenceId:  5,
			ChannelId:    1,
		},
		want: 0,
	}, {
		name: "Find corridor with FromNodeId 2",
		input: CorridorKey{
			CorridorType: Tag(),
			ReferenceId:  5,
			FromNodeId:   2,
		},
		want: 1,
	}, {
		name: "Find corridor with FromNodeId: 3",
		input: CorridorKey{
			CorridorType: Tag(),
			ReferenceId:  5,
			FromNodeId:   3,
		},
		want: 0,
	},
	}

	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := GetBestCorridorFlag(test.input)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("%d: GetBestCorridorFlag()\nGot:\n%v\nWant:\n%v\n", i, got, test.want)
			}
		})
	}
}

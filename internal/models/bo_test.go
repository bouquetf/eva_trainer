package models

import (
	"testing"
)

func ptr(n int) *int { return &n }

func TestComputeBOWinner(t *testing.T) {
	tests := []struct {
		name      string
		matchs    []*Match
		equipe1ID int
		equipe2ID int
		forfait   int
		want      *int
	}{
		{
			name: "team1 wins all",
			matchs: []*Match{
				{VainqueurID: ptr(1)},
				{VainqueurID: ptr(1)},
			},
			equipe1ID: 1,
			equipe2ID: 2,
			want:      ptr(1),
		},
		{
			name: "team2 wins all",
			matchs: []*Match{
				{VainqueurID: ptr(2)},
				{VainqueurID: ptr(2)},
				{VainqueurID: ptr(2)},
			},
			equipe1ID: 1,
			equipe2ID: 2,
			want:      ptr(2),
		},
		{
			name: "equal wins",
			matchs: []*Match{
				{VainqueurID: ptr(1)},
				{VainqueurID: ptr(2)},
			},
			equipe1ID: 1,
			equipe2ID: 2,
			want:      nil,
		},
		{
			name: "team1 leads 2-1",
			matchs: []*Match{
				{VainqueurID: ptr(1)},
				{VainqueurID: ptr(2)},
				{VainqueurID: ptr(1)},
			},
			equipe1ID: 1,
			equipe2ID: 2,
			want:      ptr(1),
		},
		{
			name: "some matches with nil winner",
			matchs: []*Match{
				{VainqueurID: ptr(1)},
				{VainqueurID: nil},
				{VainqueurID: ptr(1)},
			},
			equipe1ID: 1,
			equipe2ID: 2,
			want:      ptr(1),
		},
		{
			name:      "empty match list",
			matchs:    []*Match{},
			equipe1ID: 1,
			equipe2ID: 2,
			want:      nil,
		},
		{
			name: "all nil winners",
			matchs: []*Match{
				{VainqueurID: nil},
				{VainqueurID: nil},
			},
			equipe1ID: 1,
			equipe2ID: 2,
			want:      nil,
		},
		{
			name:      "forfait equipe1 gives win to equipe2",
			matchs:    nil,
			equipe1ID: 1,
			equipe2ID: 2,
			forfait:   1,
			want:      ptr(2),
		},
		{
			name:      "forfait equipe2 gives win to equipe1",
			matchs:    nil,
			equipe1ID: 1,
			equipe2ID: 2,
			forfait:   2,
			want:      ptr(1),
		},
		{
			name: "hors_competition match is ignored",
			matchs: []*Match{
				{VainqueurID: ptr(1), HorsCompetition: false},
				{VainqueurID: ptr(2), HorsCompetition: true},
				{VainqueurID: ptr(1), HorsCompetition: false},
			},
			equipe1ID: 1,
			equipe2ID: 2,
			want:      ptr(1),
		},
		{
			name: "all matches hors_competition => no winner",
			matchs: []*Match{
				{VainqueurID: ptr(1), HorsCompetition: true},
				{VainqueurID: ptr(2), HorsCompetition: true},
			},
			equipe1ID: 1,
			equipe2ID: 2,
			want:      nil,
		},
		{
			name: "forfait + hors_competition",
			matchs: []*Match{
				{VainqueurID: ptr(1), HorsCompetition: false},
				{VainqueurID: ptr(2), HorsCompetition: true},
			},
			equipe1ID: 1,
			equipe2ID: 2,
			want:      ptr(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeBOWinner(tt.matchs, tt.equipe1ID, tt.equipe2ID, tt.forfait)
			if tt.want == nil && got != nil {
				t.Errorf("ComputeBOWinner() = %v, want nil", *got)
				return
			}
			if tt.want != nil && got == nil {
				t.Errorf("ComputeBOWinner() = nil, want %d", *tt.want)
				return
			}
			if tt.want != nil && *got != *tt.want {
				t.Errorf("ComputeBOWinner() = %d, want %d", *got, *tt.want)
			}
		})
	}
}

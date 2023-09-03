package cache

import (
	"fmt"
	"github.com/goccy/go-json"
	"testing"

	"go.dot.industries/brease/models"
	"go.dot.industries/brease/pb"
)

var (
	things   = []string{"a", "b", "c"}
	v2, _    = json.Marshal(2)
	v4, _    = json.Marshal(4)
	ruleset1 = []models.VersionedRule{
		{
			Rule: models.Rule{
				ID:          "asdf",
				Description: "first rule",
				Actions: []models.Action{
					{
						Action: "setValue",
						Target: models.Target{
							Kind:   "jsonpath",
							Target: "$.property2",
							Value:  "newValue",
						},
					},
				},
				Expression: map[string]interface{}{
					"and": pb.And{
						Expression: []*pb.Expression{
							{
								Expr: &pb.Expression_Condition{
									Condition: &pb.Condition{
										Base:  &pb.Condition_Key{Key: "$.property3"},
										Kind:  "lt",
										Value: v2,
									},
								},
							},
							{
								Expr: &pb.Expression_Condition{
									Condition: &pb.Condition{
										Base:  &pb.Condition_Key{Key: "$.property"},
										Kind:  "gt",
										Value: v4,
									},
								},
							},
						},
					},
				},
			},
		},
	}
)

func BenchmarkSimpleArray(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SimpleHash(things)
	}
}

func BenchmarkRuleset(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SimpleHash(ruleset1)
	}
}

type Student struct {
	Name    string
	Address []string
	School  School
}

type School struct {
	Labels   map[string]any
	Teachers []Teacher
}

type Teacher struct {
	Subjects []string
}

func TestHashStruct(t *testing.T) {
	student1 := Student{
		Name:    "xiaoming",
		Address: []string{"mumbai", "london", "tokyo", "seattle"},
		School: School{
			Labels: map[string]any{
				"phone":   "123456",
				"country": "China",
			},
			Teachers: []Teacher{{Subjects: []string{"math", "chinese", "art"}}},
		},
	}

	student1UnOrder := Student{
		Name:    "xiaoming",
		Address: []string{"mumbai", "london", "seattle", "tokyo"},
		School: School{
			Labels: map[string]any{
				"phone":   "123456",
				"country": "China",
			},
			Teachers: []Teacher{{Subjects: []string{"math", "chinese", "art"}}},
		},
	}

	s1 := SimpleHash(student1)
	s2 := SimpleHash(student1UnOrder)
	if s1 != s2 {
		t.Errorf("Content order made a difference...")
	}
	msg := fmt.Sprintf("student1 hash: %s, student2 hash: %s, student1 == student2 ? -> %t", s1, s2, s1 == s2)
	fmt.Println(msg)

	student3 := Student{
		// Name is different from student1, student1UnOrder
		Name:    "xiaohong",
		Address: []string{"mumbai", "london", "seattle", "tokyo"},
		School: School{
			Labels: map[string]any{
				"phone":   "123456",
				"country": "China",
			},
			Teachers: []Teacher{{Subjects: []string{"math", "chinese", "art"}}},
		},
	}

	s3 := SimpleHash(student3)
	fmt.Println(fmt.Sprintf("student3 hash: %s, student2 == student3 ? -> %t", s3, s2 == s3))
}

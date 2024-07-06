package cache

import (
	"fmt"
	"testing"

	"github.com/goccy/go-json"

	v11 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/rule/v1"
)

var (
	things   = []string{"a", "b", "c"}
	v2, _    = json.Marshal(2)
	v4, _    = json.Marshal(4)
	ruleset1 = []v11.VersionedRule{
		{
			Id:          "asdf",
			Version:     0,
			Description: "first rule",
			Actions: []*v11.Action{
				{
					Kind: "setValue",
					Target: &v11.Target{
						Kind:  "jsonpath",
						Id:    "$.property2",
						Value: []byte("newValue"),
					},
				},
			},
			Expression: &v11.Expression{
				Expr: &v11.Expression_And{
					And: &v11.And{
						Expression: []*v11.Expression{
							{
								Expr: &v11.Expression_Condition{
									Condition: &v11.Condition{
										Base: &v11.Condition_Key{
											Key: "$.property3",
										},
										Kind:  v11.ConditionKind(v11.ConditionKind_value["lt"]),
										Value: v2,
									},
								},
							},
							{
								Expr: &v11.Expression_Condition{
									Condition: &v11.Condition{
										Base: &v11.Condition_Key{
											Key: "$.property",
										},
										Kind:  v11.ConditionKind(v11.ConditionKind_value["gt"]),
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

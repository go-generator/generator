package manual

type File struct {
	Name     string `json:"name,omitempty" bson:"name,omitempty" dynamodbav:"name,omitempty" firestore:"name,omitempty"`
	FullName string `json:"fullName,omitempty" bson:"fullName,omitempty" dynamodbav:"fullName,omitempty" firestore:"fullName,omitempty"`
	Content  string `json:"content,omitempty" bson:"content,omitempty" dynamodbav:"content,omitempty" firestore:"content,omitempty"`
}

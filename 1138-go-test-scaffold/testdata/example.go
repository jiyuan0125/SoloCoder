package testdata

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	IsActive  bool      `json:"is_active"`
	Tags      []string  `json:"tags"`
	Metadata  map[string]string `json:"metadata"`
	Profile   *Profile  `json:"profile"`
	CreatedAt string    `json:"created_at"`
}

type Profile struct {
	Bio      string   `json:"bio"`
	Website  string   `json:"website"`
	Location string   `json:"location"`
}

type Validator interface {
	Validate() error
}

func (u *User) Validate() error {
	return nil
}

func (u *User) FullName() string {
	return u.Name
}

func (u *User) GetTags() []string {
	return u.Tags
}

package decorator

import "final/internal/structs"

type PostComponent interface {
	Post() structs.Post
	Badges() []string
}

type edited struct {
	innerComponent PostComponent
}

type flagged struct {
	innerComponent PostComponent
	check func(structs.Post) bool
}

type basePost struct {
	post structs.Post
}

func NewPost(post structs.Post) *basePost {
	return &basePost{
		post: post,
	}
}

func (b *basePost) Post() structs.Post {
	return b.post
}

func (b *basePost) Badges() []string {
	return nil
}

func Edited(component PostComponent) PostComponent {
	return edited{innerComponent: component}
}

func (edit edited) Post() structs.Post {
	return edit.innerComponent.Post()
}

func (edit edited) Badges() []string {
	badges := edit.innerComponent.Badges()
	post := edit.innerComponent.Post()
	if post.UpdatedAt.After(post.CreatedAt) {
		badges = append(badges, "edited")
	}
	return badges
}

func Flagged(component PostComponent, check func(structs.Post) bool) PostComponent {
	return flagged{innerComponent: component, check: check}
}

func (flag flagged) Post() structs.Post {
	return flag.innerComponent.Post()
}

func (flag flagged) Badges() []string {
	badges := flag.innerComponent.Badges()
	if flag.check(flag.innerComponent.Post()) {
		badges = append(badges, "flagged")
	}
	return badges
}

type PostView struct {
	structs.Post

	Badges []string
}

func Build(component PostComponent) PostView {
	return PostView{Post: component.Post(), Badges: component.Badges()}
}
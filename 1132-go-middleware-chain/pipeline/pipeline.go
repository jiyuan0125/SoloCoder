package pipeline

type Pipeline struct {
	name  string
	chain *Chain
}

func NewPipeline(name string) *Pipeline {
	return &Pipeline{
		name:  name,
		chain: NewChain(),
	}
}

func (p *Pipeline) Name() string {
	return p.name
}

func (p *Pipeline) AddHandler(h Handler) {
	p.chain.AddHandler(h)
}

func (p *Pipeline) InsertHandlerAt(h Handler, position int) {
	p.chain.InsertHandlerAt(h, position)
}

func (p *Pipeline) RemoveHandler(id string) bool {
	return p.chain.RemoveHandler(id)
}

func (p *Pipeline) ReorderHandlers(ids []string) bool {
	return p.chain.ReorderHandlers(ids)
}

func (p *Pipeline) GetHandlerIDs() []string {
	return p.chain.GetHandlerIDs()
}

func (p *Pipeline) Execute(ctx *Context) *ChainExecutionResult {
	return p.chain.Execute(ctx)
}

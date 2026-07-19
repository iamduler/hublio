package domain

// Pipeline runs an ordered list of Operations against a Document. An empty Pipeline is an
// identity transform: Run returns a clone of the input unchanged, which is what lets
// Orchestration pass no rules for capabilities that have no Canonical normalization needs
// (e.g. the Fake connector's echo capability).
type Pipeline struct {
	ops []Operation
}

func NewPipeline(ops ...Operation) *Pipeline {
	return &Pipeline{ops: ops}
}

// Run applies every Operation in order against a clone of doc, so the caller's input map is
// never mutated by reference. It stops and returns the first Operation error encountered
// (mapping errors must stop execution before the Connector Runtime is invoked, docs/06 §26).
func (p *Pipeline) Run(doc Document) (Document, error) {
	current := doc.Clone()
	for _, op := range p.ops {
		var err error
		current, err = op.Apply(current)
		if err != nil {
			return nil, err
		}
	}
	return current, nil
}

package scanner

type Scanner struct {
}

func (s *Scanner) Start()
func (s *Scanner) Stop()
func (s *Scanner) GetUtxos()

// SetHeight Set a new internal scan height
func (s *Scanner) SetHeight(uint32)

// NewUtxosChan can only have one caller
// Data is only pushed through once.
// todo: should work like context.Context.Done()
func (s *Scanner) NewUtxosChan() <-chan any

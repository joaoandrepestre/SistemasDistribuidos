package safe

import(
	"sync"
	"sync/atomic"
)

// ThreadSafeInt - interface para impedir condição de corrida em variáveis globais
type ThreadSafeInt struct {
	Value int32
}

// Get - utiliza biblioteca atomic para serializar a leitura ao valor armazenado
func (t *ThreadSafeInt) Get() int {
	return int(atomic.LoadInt32(&t.Value))
}

// Set - utiliza biblioteca atomic para serializar a escrita ao valor armazenado
func (t *ThreadSafeInt) Set(v int) {
	atomic.StoreInt32(&t.Value, int32(v))
}

// IncrementAndGet - incrementa valor em 1 e retorna
func (t *ThreadSafeInt) IncrementAndGet() int {
	atomic.StoreInt32(&t.Value, int32(t.Get()+1))
	return int(atomic.LoadInt32(&t.Value))
}

// Decrement - decrementa o valor em 1
func (t *ThreadSafeInt) Decrement() {
	atomic.StoreInt32(&t.Value, int32(t.Get()-1))
}

// Increment - incrementa o valor em 1
func (t *ThreadSafeInt) Increment() {
	atomic.StoreInt32(&t.Value, int32(t.Get()+1))
}

// ThreadSafeBool - estrutura booleana com operações atômicas
type ThreadSafeBool struct {
	lock  sync.Mutex
	Value bool
}

// Toggle - mudar valor do booleano
func (t *ThreadSafeBool) Toggle() {
	t.lock.Lock()
	t.Value = !t.Value
	t.lock.Unlock()
}

// Get - pegar valor do booleano
func (t *ThreadSafeBool) Get() bool {
	var temp bool

	t.lock.Lock()
	temp = t.Value
	t.lock.Unlock()

	return temp
}

// Set - sobrescrever valor do booleano
func (t *ThreadSafeBool) Set(value bool) {
	t.lock.Lock()
	t.Value = value
	t.lock.Unlock()
}
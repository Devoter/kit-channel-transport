package channel

import "sync"

// Handler is a combination of a HandlerFunc and events channel.
type Handler struct {
	HandlerFunc HandlerFunc
	Channel     chan RequestEvent
}

// Manager connects client endpoints and handlers via Bus.
type Manager struct {
	bus          *Bus
	mx           sync.RWMutex
	wg           *sync.WaitGroup
	stopChannels []chan bool
	bufSize      int
	handlers     []Handler
}

// NewManager returns an initialized Manager instance.
func NewManager(bufSize int) *Manager {
	if bufSize < 1 {
		bufSize = 1
	}

	return &Manager{
		bus:          NewBus(),
		wg:           &sync.WaitGroup{},
		stopChannels: []chan bool{},
		bufSize:      bufSize,
		handlers:     []Handler{},
	}
}

// Bus returns internal bus instance.
func (m *Manager) Bus() *Bus {
	return m.bus
}

// Listen starts a channels listeners.
func (m *Manager) Listen() {
	handlersCount := len(m.handlers)
	m.wg.Add(handlersCount)

	for i := 0; i < handlersCount; i++ {
		stopCh := make(chan bool, 1)
		m.stopChannels = append(m.stopChannels, stopCh)

		go m.startListener(m.handlers[i].Channel, stopCh, m.handlers[i].HandlerFunc)
	}
}

// Stop stops all listeners.
func (m *Manager) Stop() {
	for _, c := range m.stopChannels {
		c <- true
	}

	m.wg.Wait()
}

// Register associates a handler function with a key. This method should be called before the listeners start.
func (m *Manager) Register(key string, h HandlerFunc) {
	defer m.mx.Unlock()

	m.mx.Lock()

	handler := Handler{
		HandlerFunc: h,
		Channel:     make(chan RequestEvent, m.bufSize),
	}

	m.handlers = append(m.handlers, handler)
	m.bus.Subscribe(key, handler.Channel)
}

func (m *Manager) startListener(eventsCh <-chan RequestEvent, stopCh <-chan bool, h HandlerFunc) {
	defer m.wg.Done()

	for {
		select {
		case <-stopCh:
			return
		case event := <-eventsCh:
			go h(&event)
		}
	}
}

package lib

import "sync"

var WaitG sync.WaitGroup = sync.WaitGroup{}
var GracefullShutdownChan chan struct{} = make(chan struct{})

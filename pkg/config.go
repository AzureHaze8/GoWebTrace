package pkg

// Options 存储所有命令行参数
type Options struct {
	TargetURL   string
	FilePath    string
	Concurrency int
	RulePath    string
	Output      string
	CertFinger  bool
	ProxyList   bool
	ProxyURL    string
}

// 负责代理加载和轮询
type Manager struct {
	proxies []string
	count   uint64
}

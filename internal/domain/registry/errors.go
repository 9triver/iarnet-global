package registry

import "errors"

var (
	// ErrDomainNotFound 域不存在
	ErrDomainNotFound = errors.New("domain not found")
	// ErrDomainAlreadyExists 域已存在
	ErrDomainAlreadyExists = errors.New("domain already exists")
	// ErrNodeNotFound 节点不存在
	ErrNodeNotFound = errors.New("node not found")
	// ErrNodeAlreadyExists 节点已存在
	ErrNodeAlreadyExists = errors.New("node already exists")
	// ErrNodeNotInDomain 节点不属于该域
	ErrNodeNotInDomain = errors.New("node not in domain")
	// ErrHeadNodeNotSet head 节点未设置
	ErrHeadNodeNotSet = errors.New("head node not set")
	// ErrHeadNodeOffline head 节点离线
	ErrHeadNodeOffline = errors.New("head node is offline")
	// ErrInvalidResourceTags 无效的资源标签
	ErrInvalidResourceTags = errors.New("invalid resource tags")
)

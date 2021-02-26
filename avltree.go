package gof

import (
	"errors"
	"fmt"
	"math"
)

type Node struct {
	Key    int64
	Value  []int
	Left   *Node
	Right  *Node
	Height int
}

func NewNode(k int64, v int) *Node {
	val := make([]int, 0, 1024)
	val = append(val, v)
	return &Node{
		Key:    k,
		Value:  val,
		Left:   nil,
		Right:  nil,
		Height: 1,
	}
}

type AVLTree struct {
	root *Node
	size int
}

//获取元素个数
func (a *AVLTree) GetSize() int {
	return a.size
}

func (a *AVLTree) GetRoot() int64 {
	if a.root == nil {
		return -1
	}
	return a.root.Key
}

//判断为空
func (a *AVLTree) IsEmpty() bool {
	return a.size == 0
}

//判断是否是一棵二分搜索树
func (a *AVLTree) IsBST() bool {
	keys := []int64{}
	a.inOrder(a.root, &keys)
	for i := 1; i < len(keys); i++ {
		if keys[i-1] > keys[i] {
			return false
		}
	}
	return true
}

func (a *AVLTree) InOrder(k int64) []int64 {
	keys := []int64{}
	if k == -1 {
		a.inOrder(a.root, &keys)
	} else {
		node := a.getNode(a.root, k)
		a.inOrder(node, &keys)
	}
	return keys
}

func (a *AVLTree) inOrder(node *Node, keys *[]int64) {
	if node == nil {
		return
	}
	a.inOrder(node.Left, keys)
	*keys = append(*keys, node.Key)
	a.inOrder(node.Right, keys)
}

func (a *AVLTree) IsBalanced() bool {
	return a.isBalanced(a.root)
}

//是否是平衡二叉树
func (a *AVLTree) isBalanced(node *Node) bool {
	if node == nil {
		return true
	}
	balanceFactor := a.getBalanceFactor(node)
	if math.Abs(float64(balanceFactor)) > 1 {
		return false
	}
	return a.isBalanced(node.Left) && a.isBalanced(node.Right)
}

//获取平衡因子
func (a *AVLTree) getBalanceFactor(node *Node) int {
	if node == nil {
		return 0
	}
	return a.getHeight(node.Left) - a.getHeight(node.Right)
}

//获取数的高度
func (a *AVLTree) getHeight(node *Node) int {
	if node == nil {
		return 0
	}
	return node.Height

}

//右旋转
func (a *AVLTree) rightRotate(y *Node) *Node {
	x := y.Left
	T3 := x.Right
	x.Right = y
	y.Left = T3
	y.Height = int(math.Max(float64(a.getHeight(y.Left)), float64(a.getHeight(y.Right)))) + 1
	x.Height = int(math.Max(float64(a.getHeight(x.Left)), float64(a.getHeight(x.Right)))) + 1
	return x
}

//左旋转
func (a *AVLTree) leftRotate(y *Node) *Node {
	x := y.Right
	T2 := x.Left
	x.Left = y
	y.Right = T2
	y.Height = int(math.Max(float64(a.getHeight(y.Left)), float64(a.getHeight(y.Right)))) + 1
	x.Height = int(math.Max(float64(a.getHeight(x.Left)), float64(a.getHeight(x.Right)))) + 1
	return x
}

func (a *AVLTree) Add(k int64, v int) {
	a.root = a.add(a.root, k, v)
}
func (a *AVLTree) add(node *Node, k int64, v int) *Node {
	if node == nil {
		a.size++
		return NewNode(k, v)
	}
	if k < node.Key {
		node.Left = a.add(node.Left, k, v)
	} else if k > node.Key {
		node.Right = a.add(node.Right, k, v)
	} else {
		node.Value = append(node.Value, v)
	}
	node.Height = 1 + int(math.Max(float64(a.getHeight(node.Left)), float64(a.getHeight(node.Right))))
	balanceFactor := a.getBalanceFactor(node)
	if balanceFactor > 1 && a.getBalanceFactor(node.Left) >= 0 {
		return a.rightRotate(node)
	}
	if balanceFactor < -1 && a.getBalanceFactor(node.Right) <= 0 {
		return a.leftRotate(node)
	}
	if balanceFactor > 1 && a.getBalanceFactor(node.Left) < 0 {
		node.Left = a.leftRotate(node.Left)
		return a.rightRotate(node)
	}
	if balanceFactor < -1 && a.getBalanceFactor(node.Right) > 0 {
		node.Right = a.rightRotate(node.Right)
		return a.leftRotate(node)
	}
	return node
}

func (a *AVLTree) getNode(node *Node, k int64) *Node {
	if node == nil {
		return nil
	}
	if k == node.Key {
		return node
	} else if k < node.Key {
		return a.getNode(node.Left, k)
	} else {
		return a.getNode(node.Right, k)
	}
}

func (a *AVLTree) Contains(k int64) bool {
	return a.getNode(a.root, k) != nil
}

func (a *AVLTree) Get(k int64) []int {
	node := a.getNode(a.root, k)
	if node == nil {
		return nil
	}
	return node.Value
}

func (a *AVLTree) GetLessThanKey(k int64) []int64 {
	return a.getLessThanKey(a.root, k)
}

func (a *AVLTree) getLessThanKey(node *Node, k int64) []int64 {
	if node==nil{
		return nil
	}
	if node.Key < k {
		keys := []int64{node.Key}
		a.inOrder(node.Left,&keys)
		return keys
	}
	return a.getLessThanKey(node.Left, k)
}


// @Author WangKan
// @Description //更新节点
// @Date 2021/2/25 18:19
// @Param
// @return
func (a *AVLTree) Set(oldkey, newkey int64, v int) error {
	//把旧节点中的值删除
	node := a.getNode(a.root, oldkey)
	if node == nil {
		err := errors.New(fmt.Sprintf("%d 为 key的节点不存在！", oldkey))
		return err
	}
	node.Value = a.deleteSliceElementByValue(node.Value, v)
	//如果就节点中没有值了，就将旧节点删除
	if len(node.Value) == 0 {
		a.Remove(oldkey)
	}
	//获取新节点，并将当前值塞入新节点中
	node = a.getNode(a.root, newkey)
	//如果没有新节点，就创建一个新节点
	if node == nil {
		a.Add(newkey, v)
		return nil
	}
	//如果有新key节点，就将当前的v加入到 新key节点的value中
	node.Value = append(node.Value, v)
	return nil
}

func (a *AVLTree) minimum(node *Node) *Node {
	if node.Left == nil {
		return node
	}
	return a.minimum(node.Left)
}

func (a *AVLTree) Remove(k int64) []int {
	node := a.getNode(a.root, k)
	if node != nil {
		a.root = a.remove(a.root, k)
		return node.Value
	}
	return nil
}

func (a *AVLTree) RemoveNodeValue(k int64,v int) error {
	node := a.getNode(a.root,k)
	if node == nil {
		err := errors.New(fmt.Sprintf("%d 为 key的节点不存在！", k))
		return err
	}
	node.Value = a.deleteSliceElementByValue(node.Value, v)
	//如果就节点中没有值了，就将旧节点删除
	if len(node.Value) == 0 {
		a.Remove(k)
		return nil
	}
	return nil
}

func (a *AVLTree) RemoveOneNodeAndChilds(k int64) bool {

	node := a.getNode(a.root, k)
	if node == nil {
		return true
	}
	keys := []int64{}
	a.inOrder(node, &keys)
	for _, key := range keys {
		a.root = a.remove(a.root, key)
	}
	return true

}

func (a *AVLTree) remove(node *Node, k int64) *Node {
	if node == nil {
		return nil
	}
	var retNode *Node
	if k < node.Key {
		node.Left = a.remove(node.Left, k)
		retNode = node
	} else if k > node.Key {
		node.Right = a.remove(node.Right, k)
		retNode = node
	} else {
		if node.Left == nil {
			rightNode := node.Right
			node.Right = nil
			a.size--
			retNode = rightNode
		} else if node.Right == nil {
			leftNode := node.Left
			node.Left = nil
			a.size--
			retNode = leftNode
		} else {
			successor := a.minimum(node.Right)
			successor.Right = a.remove(node.Right, successor.Key)
			successor.Left = node.Left
			node.Left, node.Right = nil, nil
			retNode = successor
		}
	}
	if retNode == nil {
		return nil
	}
	retNode.Height = 1 + int(math.Max(float64(a.getHeight(retNode.Left)), float64(a.getHeight(retNode.Right))))

	balanceFactor := a.getBalanceFactor(retNode)
	if balanceFactor > 1 && a.getBalanceFactor(retNode.Left) >= 0 {
		return a.rightRotate(retNode)
	}

	if balanceFactor < -1 && a.getBalanceFactor(retNode.Right) <= 0 {
		return a.leftRotate(retNode)
	}
	if balanceFactor > 1 && a.getBalanceFactor(retNode.Left) < 0 {
		node.Left = a.leftRotate(retNode.Left)
		return a.rightRotate(retNode)
	}
	if balanceFactor < -1 && a.getBalanceFactor(retNode.Right) > 0 {
		node.Right = a.rightRotate(retNode.Right)
		return a.leftRotate(retNode)
	}
	return retNode
}

// @Author WangKan
// @Description //从切片中删除一个元素
// @Date 2021/2/25 18:09
// @Param
// @return
func (a *AVLTree) deleteSliceElementByValue(s []int, v int) []int {
	j := 0
	for _, val := range s {
		if val != v {
			s[j] = val
			j++
		}
	}
	return s[:j]
}

func NewAvlTree() *AVLTree {
	return &AVLTree{
		root: nil,
		size: 0,
	}
}

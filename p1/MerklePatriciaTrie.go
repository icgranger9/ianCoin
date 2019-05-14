package p1

import (
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/sha3"
	"reflect"
	"strings"
)

type Flag_value struct {
	encoded_prefix []uint8
	value          string
}

type Node struct {
	node_type    int // 0: Null, 1: Branch, 2: Ext or Leaf
	branch_value [17]string
	flag_value   Flag_value
}

type MerklePatriciaTrie struct {
	db     map[string]Node
	Root   string
	KeyVal map[string]string
}

func (mpt *MerklePatriciaTrie) Get(key string) (string, error) {

	fmt.Println("In mpt.Get, with mpt of:", mpt.Order_nodes())

	var hexKey []uint8

	if key == "" {
		return "", errors.New("no_key_privided")
	} else {
		asciiKey := []uint8(key)

		hexKey = make([]uint8, len(asciiKey)*2)
		for c := 0; c < len(asciiKey); c++ {
			hexKey[2*c] = asciiKey[c] / 16
			hexKey[(2*c)+1] = asciiKey[c] % 16
		}

	}
	nextNode := mpt.db[mpt.Root]

	ind := len(hexKey)
	for ind >= 0 {

		switch nextNode.node_type {
		case 1:
			if ind == 0 {
				return nextNode.branch_value[16], nil

			} else {
				prefix := hexKey[0]
				hexKey = hexKey[1:]

				nextNode = mpt.db[nextNode.branch_value[prefix]]

				ind--

			}
		case 2:
			// check if parent is leaf or ext
			isLeaf := 1 < nextNode.flag_value.encoded_prefix[0]/16

			if isLeaf {
				leafKey := compact_decode(nextNode.flag_value.encoded_prefix)

				areEq := true
				for c := 0; len(hexKey) <= len(leafKey) && c < len(hexKey); c++ {
					if hexKey[c] != leafKey[c] {
						areEq = false
						break
					}
				}
				if len(hexKey) != len(leafKey) {
					areEq = false
				}

				if areEq {
					return nextNode.flag_value.value, nil
				} else {
					return "", errors.New("reached_invalid_leaf")
				}
			} else {

				extKey := compact_decode(nextNode.flag_value.encoded_prefix)
				areEq := true

				for c := 0; len(hexKey) >= len(extKey) && c < len(extKey); c++ {
					if hexKey[c] != extKey[c] {
						areEq = false
						break
					}
				}
				if len(hexKey) < len(extKey) {
					areEq = false
				}

				if areEq {
					lenExt := len(extKey)
					hexKey = hexKey[lenExt:]
					nextNode = mpt.db[nextNode.flag_value.value]

					ind = ind - lenExt
				} else {
					ind = -99
				}
			}

		default:
			return "", errors.New("invalid_node_encountered")

		}
	}

	return "", errors.New("path_not_found")
}

func (mpt *MerklePatriciaTrie) Insert(key string, new_value string) {
	//first test to see if either string is null, and convert key to hash

	var hexKey []uint8
	if key == "" || new_value == "" {
		return
	} else {
		asciiKey := []uint8(key)

		hexKey = make([]uint8, len(asciiKey)*2)
		for c := 0; c < len(asciiKey); c++ {
			hexKey[2*c] = asciiKey[c] / 16
			hexKey[(2*c)+1] = asciiKey[c] % 16
		}

	}

	//if not null, add to KeyValue
	mpt.KeyVal[key] = new_value

	// If db is empty, create leaf as root
	if len(mpt.db) == 0 {
		root := create_leaf(hexKey, new_value)

		mpt.Root = root.hash_node()
		mpt.db[root.hash_node()] = root

	} else {
		//Otherwise, much more complicated
		//create a recursive function to help do this

		mpt.Root = recursive_insert(mpt, mpt.db[mpt.Root], create_leaf(hexKey, new_value))

	}
}

func (mpt *MerklePatriciaTrie) Delete(key string) (string, error) {
	//first ensure that key is in the tree
	resGet, _ := mpt.Get(key)

	if resGet == "" {

		return "", errors.New("key_not_found")
	}

	// if the key is in the db, must remove it

	//remove from KeyValue
	delete(mpt.KeyVal, key)

	//remove from trie
	root := mpt.db[mpt.Root]
	isLeaf := root.node_type == 2 && 1 < root.flag_value.encoded_prefix[0]/16
	if isLeaf {

		//if the trie is just a leaf, wipe the whole thing out
		mpt.Initial()
		return "", nil

	} else {

		//if trie is more than just a leaf, much more complicated, must call rec delete
		var hexKey []uint8
		asciiKey := []uint8(key)

		hexKey = make([]uint8, len(asciiKey)*2)
		for c := 0; c < len(asciiKey); c++ {
			hexKey[2*c] = asciiKey[c] / 16
			hexKey[(2*c)+1] = asciiKey[c] % 16
		}
		var err error
		recursive_delete(mpt, root, create_null(hexKey, resGet))

		return "", err

	}
}

func compact_encode(hex_array []uint8) []uint8 {

	var res, tmp []uint8
	var term int

	// must first calculate flag, if the last value is 16
	// then figure out if the length is even or odd
	// and finally use both of those to calculate the flag
	var new_array []uint8
	if hex_array[len(hex_array)-1] == 16 {
		term = 1
		new_array = hex_array[:len(hex_array)-1] //removes 16 from the end
	} else {
		term = 0
		new_array = hex_array
	}

	oddlen := len(new_array) % 2

	flag := (2 * term) + oddlen
	flags := []uint8{uint8(flag)}

	//add those two to the front of the arr
	if oddlen != 0 {
		tmp = append(flags, new_array...)
	} else {
		tmp = append(flags, 0)
		tmp = append(tmp, new_array...)
	}

	//the convert the whole thing to base 16.
	for c := 0; c < len(tmp)-1; c += 2 {
		res = append(res, (tmp[c]*16)+tmp[c+1]) //need more robust conversion from hex
	}

	//more complex enxoding using hex functions, should fully implement
	//res = make([]byte, hex.DecodedLen(len(tmp)))
	//_,err := hex.Decode(res, tmp)

	if len(res) == 0 {
		fmt.Println(" ------------------------ Encoded arr is empty")
	}
	return res
}

// If Leaf, ignore 16 at the end
func compact_decode(encoded_arr []uint8) []uint8 {
	if len(encoded_arr) == 0 {
		return []uint8{}
	}
	var res []uint8

	//convert from ASCII to hex
	tmp := make([]uint8, len(encoded_arr)*2)
	for c := 0; c < len(encoded_arr); c++ {
		tmp[2*c] = encoded_arr[c] / 16
		tmp[(2*c)+1] = encoded_arr[c] % 16
	}

	//get flag, and use it to go back to original array
	flag := tmp[0]

	if flag%2 == 0 {
		res = append(res, tmp[2:]...)
	} else {
		res = append(res, tmp[1:]...)
	}

	//not needed, as per piazza
	//if flag > 1{
	//	res = append(res, 16)
	//}

	return res
}

func test_compact_encode() {

	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{1, 2, 3, 4, 5})), []uint8{1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 1, 2, 3, 4, 5})), []uint8{0, 1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 15, 1, 12, 11, 8, 16})), []uint8{0, 15, 1, 12, 11, 8}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{15, 1, 12, 11, 8, 16})), []uint8{15, 1, 12, 11, 8}))

	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{1, 6, 1})), []uint8{1, 6, 1}))

}

func (node *Node) hash_node() string {
	var str string
	switch node.node_type {
	case 0:
		str = ""
	case 1:
		str = "branch_"
		for _, v := range node.branch_value {
			str += v
		}
	case 2:
		str = node.flag_value.value
	}

	sum := sha3.Sum256([]byte(str))
	return "HashStart_" + hex.EncodeToString(sum[:]) + "_HashEnd"
}

func (node *Node) String() string {
	str := "empty string"
	switch node.node_type {
	case 0:
		str = "[Null Node]"
	case 1:
		str = "Branch["
		for i, v := range node.branch_value[:16] {
			str += fmt.Sprintf("%d=\"%s\", ", i, v)
		}
		str += fmt.Sprintf("value=%s]", node.branch_value[16])
	case 2:
		encoded_prefix := node.flag_value.encoded_prefix
		node_name := "Leaf"
		if is_ext_node(encoded_prefix) {
			node_name = "Ext"
		}
		ori_prefix := strings.Replace(fmt.Sprint(compact_decode(encoded_prefix)), " ", ", ", -1)
		str = fmt.Sprintf("%s<%v, value=\"%s\">", node_name, ori_prefix, node.flag_value.value)
	}
	return str
}

func node_to_string(node Node) string {
	return node.String()
}

func (mpt *MerklePatriciaTrie) Initial() {
	mpt.db = make(map[string]Node)
	mpt.KeyVal = make(map[string]string)
}

func is_ext_node(encoded_arr []uint8) bool {
	return encoded_arr[0]/16 < 2
}

func TestCompact() {
	test_compact_encode()
}

func (mpt *MerklePatriciaTrie) String() string {
	content := fmt.Sprintf("ROOT=%s\n", mpt.Root)
	for hash := range mpt.db {
		content += fmt.Sprintf("%s: %s\n", hash, node_to_string(mpt.db[hash]))
	}
	return content
}

func (mpt *MerklePatriciaTrie) Order_nodes() string {

	//quick fix, to add printing of empty MPTs
	if mpt.Root == ""{
		return "MPT is empty"
	}

	raw_content := mpt.String()
	content := strings.Split(raw_content, "\n")
	root_hash := strings.Split(strings.Split(content[0], "HashStart")[1], "HashEnd")[0]
	queue := []string{root_hash}
	i := -1
	rs := ""
	cur_hash := ""
	for len(queue) != 0 {
		last_index := len(queue) - 1
		cur_hash, queue = queue[last_index], queue[:last_index]
		i += 1
		line := ""
		for _, each := range content {
			if strings.HasPrefix(each, "HashStart"+cur_hash+"HashEnd") {
				line = strings.Split(each, "HashEnd: ")[1]
				rs += each + "\n"
				rs = strings.Replace(rs, "HashStart"+cur_hash+"HashEnd", fmt.Sprintf("Hash%v", i), -1)
			}
		}
		temp2 := strings.Split(line, "HashStart")
		flag := true
		for _, each := range temp2 {
			if flag {
				flag = false
				continue
			}
			queue = append(queue, strings.Split(each, "HashEnd")[0])
		}
	}
	return rs
}

// ------------------------------ My helper functions  ------------------------------

//note: should this be called on the mpt, or just a function called inside the original insert?
func recursive_insert(mpt *MerklePatriciaTrie, parentNode Node, childNode Node) string {

	//switch to handle each node type in a new way
	switch parentNode.node_type {
	case 0:
		//hopefully we never insert onto a null node, if we do, there's a major problem
		return "ERR: inserting onto null node"

	case 1:
		// parent is a branch node
		decodedPrefixChild := compact_decode(childNode.flag_value.encoded_prefix)

		//NOTE: do i need to check that the prefix is not empty?
		valueInBranch := parentNode.branch_value[decodedPrefixChild[0]]

		if valueInBranch == "" {

			// no collision with another node, simply ensure it is a leaf, and add to the branch

			var prefix uint8
			var remainder []uint8
			if len(decodedPrefixChild) > 0 {
				prefix, remainder = decodedPrefixChild[0], decodedPrefixChild[1:]
			} else {
				prefix, remainder = 64, []uint8{} //make cleaner
			}
			//double check that the child is a leaf, and update encoded prefix
			if childNode.flag_value.encoded_prefix[0]/16 > 1 {
				childNode.flag_value.encoded_prefix = compact_encode(append(remainder, 16))
			} else if len(remainder) != 0 {
				childNode.flag_value.encoded_prefix = compact_encode(remainder)
			}

			if childNode.flag_value.encoded_prefix[0]/16 < 2 && len(remainder) == 0 {
				//if the child is an empty ext, move one down

				extChild := mpt.db[childNode.flag_value.value]
				parentNode.branch_value[prefix] = extChild.hash_node()

			} else {
				//otherwise, add the child itself

				parentNode.branch_value[prefix] = childNode.hash_node()
			}

			//update / add branch's value in db ?
			mpt.db[parentNode.hash_node()] = parentNode
			mpt.db[childNode.hash_node()] = childNode

			return parentNode.hash_node()

		} else {
			//there is collision with another node, check how much, to see if a branch or extension must be made

			//get the node that we are overlapping with
			nodeInBranch := mpt.db[valueInBranch]

			//remove first element from child
			//note, could probably be done before if / else, since done in both
			var prefix uint8
			var remainder []uint8
			if len(decodedPrefixChild) > 0 {
				prefix, remainder = decodedPrefixChild[0], decodedPrefixChild[1:]
			} else {
				prefix, remainder = 64, []uint8{} //make cleaner
			}

			//double check that the child is a leaf, and update encoded prefix
			if childNode.flag_value.encoded_prefix[0]/16 > 1 {
				childNode.flag_value.encoded_prefix = compact_encode(append(remainder, 16))
			} else {
				childNode.flag_value.encoded_prefix = compact_encode(remainder)
			}

			//inserts child below recursively, and updates it's value in the branch
			parentNode.branch_value[prefix] = recursive_insert(mpt, nodeInBranch, childNode)

			mpt.db[parentNode.hash_node()] = parentNode //adds branch to db again, since one of it's values changed
			return parentNode.hash_node()

		}

	case 2:
		// check if parent is leaf or ext
		isLeaf := 1 < parentNode.flag_value.encoded_prefix[0]/16

		if isLeaf {
			decodedPrefixParent := compact_decode(parentNode.flag_value.encoded_prefix)
			decodedPrefixChild := compact_decode(childNode.flag_value.encoded_prefix)

			// check if either prefix is empty.
			if len(decodedPrefixParent) == 0 || len(decodedPrefixChild) == 0 {
				//must deal with if parent or is empty

				if len(decodedPrefixParent) == 0 && len(decodedPrefixChild) == 0 {

					parentNode.flag_value.value = childNode.flag_value.value
					parentHash := parentNode.hash_node()

					mpt.db[parentHash] = parentNode
					return parentHash

				} else if len(decodedPrefixParent) == 0 {

					//branch value comes from parent
					newBranch := create_branch(parentNode.flag_value.value)
					branchHash := recursive_insert(mpt, newBranch, childNode)

					return branchHash
				} else {
					//branch value comes from child
					//not actually sure if possible, but will add to be safe
					newBranch := create_branch(childNode.flag_value.value)
					branchHash := recursive_insert(mpt, newBranch, parentNode)

					return branchHash
				}

			} else if decodedPrefixParent[0] == decodedPrefixChild[0] {
				//check if there is overlap between the two nodes, if so must create extension node

				//see how much overlap there is
				overlap := 0
				for ; len(decodedPrefixChild) > overlap && len(decodedPrefixParent) > overlap; overlap++ {
					if decodedPrefixChild[overlap] != decodedPrefixParent[overlap] {
						break
					}
				}

				if len(decodedPrefixParent) == overlap && len(decodedPrefixChild) == overlap {

					parentNode.flag_value.value = childNode.flag_value.value
					parentHash := parentNode.hash_node()

					mpt.db[parentHash] = parentNode
					return parentHash
				}

				sharedNibs := decodedPrefixParent[:overlap]

				//create ext
				newExt := create_ext(sharedNibs)

				//remove shared nibbles from both nodes
				decodedPrefixParent = decodedPrefixParent[overlap:]
				decodedPrefixChild = decodedPrefixChild[overlap:]

				//double check that the child  and parent are leaves, and update encoded prefix
				if parentNode.flag_value.encoded_prefix[0]/16 > 1 {
					parentNode.flag_value.encoded_prefix = compact_encode(append(decodedPrefixParent, 16))
				} else {
					parentNode.flag_value.encoded_prefix = compact_encode(decodedPrefixParent)
				}

				if childNode.flag_value.encoded_prefix[0]/16 > 1 {
					childNode.flag_value.encoded_prefix = compact_encode(append(decodedPrefixChild, 16))
				} else {
					childNode.flag_value.encoded_prefix = compact_encode(decodedPrefixChild)
				}

				//recursively call insert, to create a branch
				newExt.flag_value.value = recursive_insert(mpt, parentNode, childNode)

				mpt.db[newExt.hash_node()] = newExt
				return newExt.hash_node()

			} else {
				//must create branch node
				branch := create_branch("")

				//insert branch into db
				branchHash := branch.hash_node()
				mpt.db[branchHash] = branch

				// insert both nodes below branch node
				branchHash = recursive_insert(mpt, mpt.db[branchHash], parentNode)
				branchHash = recursive_insert(mpt, mpt.db[branchHash], childNode)

				////add leaves to db
				//mpt.db[parentNode.hash_node()] = parentNode
				//mpt.db[childNode.hash_node()] = childNode

				//return hash of branch node
				return branchHash

			}

		} else {
			//then its an ext, must be treated differently

			//check how much ext and child overlap
			decodedPrefixParent := compact_decode(parentNode.flag_value.encoded_prefix)
			decodedPrefixChild := compact_decode(childNode.flag_value.encoded_prefix)

			overlap := 0
			for ; len(decodedPrefixChild) > overlap && len(decodedPrefixParent) > overlap; overlap++ {
				if decodedPrefixChild[overlap] != decodedPrefixParent[overlap] {
					break
				}
			}

			//create branch, and add both below it
			// know there will be no conflict because overlap is 0
			if overlap == 0 {
				//no overlap between ext and other node

				//three cases
				//parent (ext) is empty
				//child is empty
				//neither are empty

				var newBranch Node
				var branchHash string

				//in case either is now 0 long
				if len(decodedPrefixParent) == 0 {
					//double check that the child  and parent are leaves, and update encoded prefix
					if childNode.flag_value.encoded_prefix[0]/16 > 1 {
						childNode.flag_value.encoded_prefix = compact_encode(append(decodedPrefixChild, 16))
					} else {
						childNode.flag_value.encoded_prefix = compact_encode(decodedPrefixChild)
					}

					newBranch = create_branch("") // "" because ext can have no value

					branchHash = newBranch.hash_node()

					mpt.db[branchHash] = newBranch

					branchHash = recursive_insert(mpt, mpt.db[branchHash], childNode)

				} else if len(decodedPrefixChild) == 0 {

					//double check that the child  and parent are leaves, and update encoded prefix
					if parentNode.flag_value.encoded_prefix[0]/16 > 1 {
						parentNode.flag_value.encoded_prefix = compact_encode(append(decodedPrefixParent, 16))
					} else {
						parentNode.flag_value.encoded_prefix = compact_encode(decodedPrefixParent)
					}

					newBranch = create_branch(childNode.flag_value.value)

					branchHash = newBranch.hash_node()

					mpt.db[branchHash] = newBranch

					branchHash = recursive_insert(mpt, mpt.db[branchHash], parentNode)

				} else {

					//double check that the child  and parent are leaves, and update encoded prefix
					if parentNode.flag_value.encoded_prefix[0]/16 > 1 {
						parentNode.flag_value.encoded_prefix = compact_encode(append(decodedPrefixParent, 16))
					} else {
						parentNode.flag_value.encoded_prefix = compact_encode(decodedPrefixParent)
					}

					if childNode.flag_value.encoded_prefix[0]/16 > 1 {
						childNode.flag_value.encoded_prefix = compact_encode(append(decodedPrefixChild, 16))
					} else {
						childNode.flag_value.encoded_prefix = compact_encode(decodedPrefixChild)
					}

					newBranch = create_branch("")

					branchHash = newBranch.hash_node()

					mpt.db[branchHash] = newBranch

					branchHash = recursive_insert(mpt, mpt.db[branchHash], childNode)
					branchHash = recursive_insert(mpt, mpt.db[branchHash], parentNode)
				}

				return branchHash

			} else if len(decodedPrefixParent) == overlap {
				//if they overlap completely, remove prefix from child, and add child to branch under ext

				decodedPrefixChild = decodedPrefixChild[overlap:]

				if len(decodedPrefixChild) == 0 {
					//empty prefix for child, add it's value to branch[16]
					branch := mpt.db[parentNode.flag_value.value]

					branch.branch_value[16] = childNode.flag_value.value

					mpt.db[branch.hash_node()] = branch
					parentNode.flag_value.value = branch.hash_node()

				} else {
					//must add child to branch

					//double check that the child is a leaf, and update encoded prefix
					if childNode.flag_value.encoded_prefix[0]/16 > 1 {
						childNode.flag_value.encoded_prefix = compact_encode(append(decodedPrefixChild, 16))
					} else {
						childNode.flag_value.encoded_prefix = compact_encode(decodedPrefixChild)
					}
					//update the hash so it points to the correct node
					branchHash := recursive_insert(mpt, mpt.db[parentNode.flag_value.value], childNode)
					parentNode.flag_value.value = branchHash
				}

				//add ext node to db, and return the hash
				mpt.db[parentNode.hash_node()] = parentNode
				return parentNode.hash_node()

			} else {
				//if only partially, shrink ext, add branch, add branch under ext to new branch, and add child to new branch
				//must also update prefixes of both

				newExt := create_ext(decodedPrefixParent[:overlap])

				//remove shared nibbles from both nodes
				decodedPrefixParent = decodedPrefixParent[overlap:]
				decodedPrefixChild = decodedPrefixChild[overlap:]

				var newBranch Node
				var branchHash string

				//in case either is now 0 long
				if len(decodedPrefixParent) == 0 {
					//double check that the child  and parent are leaves, and update encoded prefix
					if childNode.flag_value.encoded_prefix[0]/16 > 1 {
						childNode.flag_value.encoded_prefix = compact_encode(append(decodedPrefixChild, 16))
					} else {
						childNode.flag_value.encoded_prefix = compact_encode(decodedPrefixChild)
					}

					newBranch = create_branch(parentNode.flag_value.value) //"" because ext doesn't have a real value'

					branchHash = newBranch.hash_node()

					mpt.db[branchHash] = newBranch

					branchHash = recursive_insert(mpt, mpt.db[branchHash], childNode)

				} else if len(decodedPrefixChild) == 0 {

					//double check that the child  and parent are leaves, and update encoded prefix
					if parentNode.flag_value.encoded_prefix[0]/16 > 1 {
						parentNode.flag_value.encoded_prefix = compact_encode(append(decodedPrefixParent, 16))
					} else {
						parentNode.flag_value.encoded_prefix = compact_encode(decodedPrefixParent)
					}

					newBranch = create_branch(childNode.flag_value.value)

					branchHash = newBranch.hash_node()

					mpt.db[branchHash] = newBranch

					branchHash = recursive_insert(mpt, mpt.db[branchHash], parentNode)

				} else {
					//double check that the child  and parent are leaves, and update encoded prefix
					if parentNode.flag_value.encoded_prefix[0]/16 > 1 {
						parentNode.flag_value.encoded_prefix = compact_encode(append(decodedPrefixParent, 16))
					} else {
						parentNode.flag_value.encoded_prefix = compact_encode(decodedPrefixParent)
					}

					if childNode.flag_value.encoded_prefix[0]/16 > 1 {
						childNode.flag_value.encoded_prefix = compact_encode(append(decodedPrefixChild, 16))
					} else {
						childNode.flag_value.encoded_prefix = compact_encode(decodedPrefixChild)
					}

					newBranch = create_branch("")

					branchHash = newBranch.hash_node()

					mpt.db[branchHash] = newBranch

					branchHash = recursive_insert(mpt, mpt.db[branchHash], childNode)
					branchHash = recursive_insert(mpt, mpt.db[branchHash], parentNode)
				}

				newExt.flag_value.value = branchHash

				mpt.db[newExt.hash_node()] = newExt

				return newExt.hash_node()

			}

		}
	default:
		//what should the default be

	}

	return parentNode.hash_node() //is this really the default has we should have?
}

func recursive_delete(mpt *MerklePatriciaTrie, parentNode Node, childNode Node) string {

	switch parentNode.node_type {
	case 0:
		//hopefully never a null node
		fmt.Println("---------------------------------> NULL NODE Encountered in delete")
	case 1:
		//if value of branch, delete it

		if childNode.node_type == 0 {
			numInBranch := 0

			//note should compare prefix, not value
			if len(compact_decode(childNode.flag_value.encoded_prefix)) == 0 {
				originalHash := parentNode.hash_node()
				parentNode.branch_value[16] = ""

				parentOfBranch := update_parent(mpt, mpt.db[originalHash], parentNode.hash_node())
				mpt.db[parentNode.hash_node()] = parentNode

				//check if we need to change the branch to something else
				var branchChild Node
				var branchVal uint8
				for c := 0; c < len(parentNode.branch_value); c++ {
					if parentNode.branch_value[c] != "" {
						branchChild = mpt.db[parentNode.branch_value[c]]
						branchVal = uint8(c)
						numInBranch++
					}
				}

				if numInBranch == 1 {
					shrink_branch(mpt, parentOfBranch, branchChild, parentNode, branchVal)
				}
				//otherwise, we're done
				return parentOfBranch.hash_node()

			} else {
				//gets prefix from child, and removes the first
				decodedChildPrefix := compact_decode(childNode.flag_value.encoded_prefix)
				nextVal, remainder := decodedChildPrefix[0], decodedChildPrefix[1:]

				//re-encodes the child, and gets the correct next node
				childNode.flag_value.encoded_prefix = compact_encode(append(remainder, 16)) //technically shouldn't add 16 since it's not a leaf, but it needs something


				nextNode := mpt.db[parentNode.branch_value[nextVal]] //note, do we need to double check that the child is actually in the branch, or can we assume


				newParent := mpt.db[recursive_delete(mpt, nextNode, childNode)]

				parentOfBranch := update_parent(mpt, newParent, newParent.hash_node())
				var branchChild Node
				var branchVal uint8

				for c := 0; c < len(newParent.branch_value); c++ {
					if newParent.branch_value[c] != "" {
						branchChild = mpt.db[newParent.branch_value[c]]
						branchVal = uint8(c)
						numInBranch++
					}
				}

				if numInBranch == 1 {
					newNode := shrink_branch(mpt, parentOfBranch, branchChild, newParent, branchVal)

					return newNode.hash_node()
				}

				return newParent.hash_node()
			}

		}

		//if only one in branch:

		//may convert to leaf if it has a value
		//would them have to merge that leaf if there was an ext above

		//if child node is below the branch, simply pass it along
	case 2:
		isLeaf := 1 < parentNode.flag_value.encoded_prefix[0]/16

		if isLeaf {

			//if it is a leaf couldn't we just assume that we found the correct one
			areEq := true
			for c := 0; len(parentNode.flag_value.value) == len(childNode.flag_value.value) && c < len(parentNode.flag_value.value); c++ {
				if parentNode.flag_value.value[c] != childNode.flag_value.value[c] {
					areEq = false
					break
				}
			}

			if areEq {
				newParent := update_parent(mpt, parentNode, "")

				mpt.db[newParent.hash_node()] = newParent

				return newParent.hash_node()
			}
		} else {
			//gets prefix from child, and removes the first
			decodedChildPrefix := compact_decode(childNode.flag_value.encoded_prefix)
			decodedParentPrefix := compact_decode(parentNode.flag_value.encoded_prefix)

			//note: prefix incorrect,, but no point if fixing
			var remainder []uint8
			if len(decodedChildPrefix) == len(decodedParentPrefix) {
				remainder = []uint8{}
			} else {
				_, remainder = decodedChildPrefix[len(decodedParentPrefix)], decodedChildPrefix[len(decodedParentPrefix):]
			}

			//re-encodes the child, and gets the correct next node
			childNode.flag_value.encoded_prefix = compact_encode(append(remainder, 16))
			nextNode := mpt.db[parentNode.flag_value.value] //note, do we need to double check that the child is actually in the branch, or can we assume

			newParent := mpt.db[recursive_delete(mpt, nextNode, childNode)]

			return newParent.hash_node()
		}
	default:
		//hopefully never called

	}

	return ""
}

func shrink_branch(mpt *MerklePatriciaTrie, branchParent Node, branchChild Node, branch Node, branchVal uint8) Node {
	//cases:
	//  null - null - branch 	: convert to ext
	//  null - null - leaf 		: add to front
	//branch - null -  null 	: get val in branch, and convert to leaf
	//branch - null -  branch 	: convert to ext
	//branch - null - ext 		: add to ext
	//branch - null - leaf 		: add to leaf
	//   ext - null - null		: get val in branch, and convert to leaf
	//   ext - null - branch	: add to ext
	//   ext - null - ext 		: merge ext's
	//   ext - null - leaf 		: merge everything


	//used in update_parent at bottom
	originalParentHash := branchParent.hash_node()
	var res Node

	switch branchParent.node_type {
	case 0:

		originalParentHash = mpt.Root

		if branchChild.node_type == 1 {
			ext := create_ext([]uint8{branchVal})
			ext.flag_value.value = branchChild.hash_node()

			mpt.db[ext.hash_node()] = ext
			res = ext
		} else {
			decodedPrefix := compact_decode(branchChild.flag_value.encoded_prefix)
			newPrefix := append([]uint8{branchVal}, decodedPrefix...)
			newPrefix = compact_encode(append(newPrefix, 16))

			branchChild.flag_value.encoded_prefix = newPrefix

			mpt.db[branchChild.hash_node()] = branchChild
			res = branchChild
		}

	case 1:
		//finds the index in branch_value that must be update
		var branchInd int
		for c := 0; c < len(branchParent.branch_value)-1; c++ {
			if branchParent.branch_value[c] == branch.hash_node() {
				branchInd = c
			}
		}
		if branchChild.node_type == 0 {
			leaf := create_leaf([]uint8{}, branch.branch_value[16])

			branchParent.branch_value[branchInd] = leaf.hash_node()

			mpt.db[branchParent.hash_node()] = branchParent

			res = branchParent

		} else if branchChild.node_type == 1 {

			ext := create_ext([]uint8{branchVal})
			ext.flag_value.value = branchChild.hash_node()

			branchParent.branch_value[branchInd] = ext.hash_node()

			mpt.db[branchParent.hash_node()] = branchParent

			res = branchParent
		} else {

			isLeaf := 1 < branchChild.flag_value.encoded_prefix[0]/16

			if isLeaf {
				decodedPrefix := compact_decode(branchChild.flag_value.encoded_prefix)
				newPrefix := append([]uint8{branchVal}, decodedPrefix...)
				newPrefix = compact_encode(append(newPrefix, 16))

				branchChild.flag_value.encoded_prefix = newPrefix

				branchParent.branch_value[branchInd] = branchChild.hash_node()

				mpt.db[branchParent.hash_node()] = branchParent
				mpt.db[branchChild.hash_node()] = branchChild
				res = branchParent
			} else {
				decodedPrefix := compact_decode(branchChild.flag_value.encoded_prefix)
				newPrefix := append([]uint8{branchVal}, decodedPrefix...)
				newPrefix = compact_encode(newPrefix)

				branchChild.flag_value.encoded_prefix = newPrefix

				branchParent.branch_value[branchInd] = branchChild.hash_node()

				mpt.db[branchParent.hash_node()] = branchParent
				mpt.db[branchChild.hash_node()] = branchChild
				res = branchParent
			}

		}

	case 2:

		if branchChild.node_type == 0 {

			extPrefix := compact_decode(branchParent.flag_value.encoded_prefix)
			leaf := create_leaf(extPrefix, branch.branch_value[16])

			mpt.db[leaf.hash_node()] = leaf

			res = leaf

		} else if branchChild.node_type == 1 {

			decodedPrefix := compact_decode(branchParent.flag_value.encoded_prefix)
			newPrefix := append([]uint8{branchVal}, decodedPrefix...)
			newPrefix = compact_encode(newPrefix)

			branchParent.flag_value.encoded_prefix = newPrefix
			branchParent.flag_value.value = branchChild.hash_node()

			mpt.db[branchParent.hash_node()] = branchParent
			res = branchParent
		} else {

			isLeaf := 1 < branchChild.flag_value.encoded_prefix[0]/16

			if isLeaf {
				decodedPrefixParent := compact_decode(branchParent.flag_value.encoded_prefix)
				decodedPrefixChild := compact_decode(branchChild.flag_value.encoded_prefix)

				newPrefix := append(decodedPrefixParent, branchVal)
				newPrefix = append(newPrefix, decodedPrefixChild...)
				newPrefix = compact_encode(append(newPrefix, 16))

				branchChild.flag_value.encoded_prefix = newPrefix

				mpt.db[branchChild.hash_node()] = branchChild
				res = branchChild

			} else {

				decodedPrefixParent := compact_decode(branchParent.flag_value.encoded_prefix)
				decodedPrefixChild := compact_decode(branchChild.flag_value.encoded_prefix)

				newPrefix := append(decodedPrefixParent, branchVal)
				newPrefix = append(newPrefix, decodedPrefixChild...)
				newPrefix = compact_encode(newPrefix)

				branchParent.flag_value.encoded_prefix = newPrefix
				branchParent.flag_value.value = branchChild.flag_value.value

				mpt.db[branchParent.hash_node()] = branchParent
				res = branchParent
			}

		}
	}

	//after switch, add to tree correctly
	update_parent(mpt, mpt.db[originalParentHash], res.hash_node())

	return res
}

func update_parent(mpt *MerklePatriciaTrie, childNode Node, newChildHash string) Node {
	//updates the hash value of the parent, and all it's parents

	//update root when needed
	var null Node

	if mpt.Root == childNode.hash_node() {
		mpt.Root = newChildHash

		return null
	}

	//compare next values of all nodes in db to the hash of the child
	//very inefficent to repeatedly whit huge trie, must be a better way

	nodes := get_nodes(mpt, mpt.db[mpt.Root])
	for _, node := range nodes {
		//handle if branch
		if node.node_type == 1 {
			for c := 0; c < len(node.branch_value)-1; c++ {

				if node.branch_value[c] == childNode.hash_node() {
					originalHash := node.hash_node()

					node.branch_value[c] = newChildHash

					mpt.db[node.hash_node()] = node

					update_parent(mpt, mpt.db[originalHash], node.hash_node())
					return node
				}
			}
		}

		//handle if ext
		if node.flag_value.value == childNode.hash_node() {
			originalHash := node.hash_node()

			node.flag_value.value = newChildHash

			mpt.db[node.hash_node()] = node
			update_parent(mpt, mpt.db[originalHash], node.hash_node())
			return node
		}
	}

	//if no parent found
	return null

}

func get_nodes(mpt *MerklePatriciaTrie, node Node) []Node {
	var res []Node
	res = append(res, node)

	switch node.node_type {
	case 1:
		for c := 0; c < len(node.branch_value)-1; c++ {
			if node.branch_value[c] != "" {
				child := mpt.db[node.branch_value[c]]
				res = append(res, get_nodes(mpt, child)...)
			}
		}
	case 2:
		child := mpt.db[node.flag_value.value]

		if child.node_type != 0 {
			res = append(res, get_nodes(mpt, child)...)
		}
	}

	return res
}

func create_leaf(hexKey []uint8, value string) Node {
	var newLeaf Node

	newLeaf.node_type = 2
	newLeaf.flag_value.encoded_prefix = compact_encode(append(hexKey, 16))
	newLeaf.flag_value.value = value

	return newLeaf
}

func create_branch(value string) Node {
	//not much needed in this function, but still cteated / used for consistency

	var newBranch Node

	newBranch.node_type = 1
	newBranch.branch_value[16] = value //in  slot 16, not in flag_value.value
	newBranch.flag_value.encoded_prefix = []uint8{}

	return newBranch
}

func create_ext(shared []uint8) Node {
	var newExt Node

	newExt.node_type = 2
	newExt.flag_value.encoded_prefix = compact_encode(shared)

	return newExt
}

func create_null(hexKey []uint8, value string) Node {
	var newNull Node

	newNull.node_type = 0
	newNull.flag_value.encoded_prefix = compact_encode(append(hexKey, 16))
	newNull.flag_value.value = value

	return newNull
}

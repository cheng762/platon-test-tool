package testcases

import (
	"errors"
	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/crypto/sha3"
	"github.com/PlatONnetwork/PlatON-Go/rlp"
)


type govCases struct {
	commonCases
}


func (r *govCases) Prepare() error {
	if err := r.commonCases.Prepare(); err != nil {
		return err
	}
	return nil
}

func (r *govCases) Start() error {

	return nil
}

func (r *govCases) Exec(caseName string) error {
	return r.commonCases.exec(caseName, r)
}

func (r *govCases) End() error {
	if err := r.commonCases.End(); err != nil {
		return err
	}
	if len(r.errors) != 0 {
		for _, value := range r.errors {
			r.SendError("gov", value)
		}
		return errors.New("run govCases fail")
	}
	return nil
}

func (r *govCases) List() []string {
	return r.list(r)
}



func (r *govCases) CaseDeclareVersion() error {
	//ctx := context.Background()
	//node1,err:= NewCommonCases("http://127.0.0.1:6771")
	//if err!=nil{
	//	return err
	//}
	//node2,err:=NewCommonCases("http://127.0.0.1:6772")
	//if err!=nil{
	//	return err
	//}
	//node3,err:= NewCommonCases("http://127.0.0.1:6773")
	//if err!=nil{
	//	return err
	//}
	//account:= new(PriAccount)
	//account.Address = common.HexToAddress("0xC7d5b0261ce3FC6C89a94e241632cae21DCfF9F8")
	//account.Priv = crypto.HexMustToECDSA("a5397318896190712f49eca2a13c1a372566d5efffb9841f1b6892e7023907bb")
	//
	//
	//nodeInfo1:=  discover.MustParseNode("enode://51f0936a365d3018b96cd221208497661e1d6664da20c3b85f27b7c08e11ebc38f042574d7e3636b9c9225f540a745ae71d4c5de816cf7770e8fb1a4f0fa28c3@127.0.0.1:16789")
	//nodeInfo2:=  discover.MustParseNode("enode://e7c57fb51cc21a1eb3f5c5933c8c3c63b6239be1d2504c1a0b4752ed40d6b34c10db80114480865662aabf1f2704602916598ba5dbf31560d1a13d942c818f00@127.0.0.1:16790")
	//nodeInfo3:=  discover.MustParseNode("enode://2e9507ea69bcf58e6c6e0fcb73299ef5774e2737d7df5ae04c4a04869742b8450313017f2dc246fcd57ba92730b7feb9ff4c057504458b25e68c8f22a1e1146e@127.0.0.1:16791")
	//
	//
	//p1:= crypto.HexMustToECDSA("4551b78fd05ef46721a51190ba883c4d695f8222a8e04263296327e04b512f0a")
	//p2:= crypto.HexMustToECDSA("a6b6f04ff55647a84f7bd9dea047e88bc8c47eee6cc4af683282d6c0e22bbe8c")
	//p3:= crypto.HexMustToECDSA("361252d1f93765e53b05a0ce8340e70b416902dc53891cb822f9e8ebdf07c691")
	//
	//hash1,err:= node1.DeclareVersion(ctx,account,p1,nodeInfo1.ID,plugin.FORKVERSION)
	//if err!=nil{
	//	return err
	//}
	//hash2,err:= node2.DeclareVersion(ctx,account,p2,nodeInfo2.ID,plugin.FORKVERSION)
	//if err!=nil{
	//	return err
	//}
	//
	//hash3,err:= node3.DeclareVersion(ctx,account,p3,nodeInfo3.ID,plugin.FORKVERSION)
	//if err!=nil{
	//	return err
	//}




	//log.Printf("finish DeclareVersion,hash1:%v,hash2:%v,hash3:%v",hash1.String(),hash2.String(),hash3.String())
	return nil
}


func RlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}
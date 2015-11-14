package commands

import (
	"testing"

	. "github.com/karlseguin/expect"
)

type CommandsTests struct {
}

func Test_Commands(t *testing.T) {
	Expectify(new(CommandsTests), t)
}

func (_ CommandsTests) LabelUint() {
	Expect(labelUint("over", 9000)).To.Eql("over: 9000")
	Expect(labelUint("power", 0)).To.Eql("power: 0")
}

func (_ CommandsTests) ReableBytes() {
	Expect(readableBytes(0)).To.Equal("0B")
	Expect(readableBytes(512)).To.Equal("512B")
	Expect(readableBytes(1023)).To.Equal("1023B")
	Expect(readableBytes(1024)).To.Equal("1KB")
	Expect(readableBytes(1048575)).To.Equal("1023KB")
	Expect(readableBytes(1048576)).To.Equal("1MB")
	Expect(readableBytes(1073741824)).To.Equal("1GB")
	Expect(readableBytes(5368709120)).To.Equal("5GB")
	Expect(readableBytes(1099511627776)).To.Equal("1TB")
	Expect(readableBytes(272678883688448)).To.Equal("248TB")
	Expect(readableBytes(2251799813685248)).To.Equal("2PB")
}

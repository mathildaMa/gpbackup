package helper_test

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/greenplum-db/gpbackup/helper"
	"github.com/greenplum-db/gpbackup/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("helper/helper", func() {
	var stdout *gbytes.Buffer
	var stdinRead, stdinWrite *os.File
	var tocFileRead, tocFileWrite *os.File
	BeforeEach(func() {
		stdinRead, stdinWrite, _ = os.Pipe()
		tocFileRead, tocFileWrite, _ = os.Pipe()
		utils.System.Stdin = stdinRead
		stdout = gbytes.NewBuffer()
		utils.System.Stdout = stdout
	})
	AfterEach(func() {
		utils.System.OpenFileRead = utils.OpenFileRead
		utils.System.Stat = os.Stat
		utils.System.Stdin = os.Stdin
		utils.System.Stdout = os.Stdout
		utils.System.ReadFile = ioutil.ReadFile
	})
	Describe("ReadAndCountBytes", func() {
		It("Returns correct number of bytes read", func() {
			fmt.Fprintln(stdinWrite, "some text")
			stdinWrite.Close()
			bytesRead := helper.ReadAndCountBytes()
			Expect(bytesRead).To(Equal(uint64(10)))
			Expect(stdout).To(gbytes.Say("some text\n"))
		})
		It("Returns 0 if no bytes read", func() {
			stdinWrite.Close()
			bytesRead := helper.ReadAndCountBytes()
			Expect(bytesRead).To(Equal(uint64(0)))
			Expect(stdout).To(gbytes.Say(""))
		})
		Describe("ReadOrCreateTOC", func() {
			It("returns contents of TOC when a TOC file exists", func() {
				helper.SetFilename("filename")
				utils.System.Stat = func(name string) (os.FileInfo, error) {
					return nil, nil
				}
				utils.System.OpenFileRead = func(name string, flag int, perm os.FileMode) (utils.ReadCloserAt, error) { return tocFileRead, nil }
				utils.System.ReadFile = func(filename string) ([]byte, error) {
					return []byte(`globalentries: []
predataentries: []
postdataentries: []
statisticsentries: []
masterdataentries: []
segmentdataentries:
- oid: 1
  startbyte: 0
  endbyte: 5
- oid: 2
  startbyte: 5
  endbyte: 10
- oid: 3
  startbyte: 10
  endbyte: 15`), nil
				}
				expectedSegmentDataEntries := []utils.SegmentDataEntry{
					{Oid: 1, StartByte: 0, EndByte: 5},
					{Oid: 2, StartByte: 5, EndByte: 10},
					{Oid: 3, StartByte: 10, EndByte: 15},
				}
				toc, lastRead := helper.ReadOrCreateTOC()
				Expect(lastRead).To(BeNumerically("==", 15))
				Expect((*toc).SegmentDataEntries).To(Equal(expectedSegmentDataEntries))
			})
			It("returns a new TOC when no TOC file exists", func() {
				helper.SetFilename("filename")
				utils.System.Stat = func(name string) (os.FileInfo, error) {
					return nil, os.ErrNotExist
				}
				toc, lastRead := helper.ReadOrCreateTOC()
				Expect(lastRead).To(Equal(uint64(0)))
				Expect((*toc).SegmentDataEntries).To(BeNil())
			})
		})
		Describe("GetBoundsForTable", func() {
			It("returns the start and end byte from the TOC", func() {
				helper.SetFilename("filename")
				helper.SetIndex(2)
				toc := &utils.TOC{}
				toc.SegmentDataEntries = []utils.SegmentDataEntry{
					{Oid: 1, StartByte: 0, EndByte: 5},
					{Oid: 2, StartByte: 5, EndByte: 10},
					{Oid: 3, StartByte: 10, EndByte: 15},
				}
				startByte, endByte := helper.GetBoundsForTable(toc)
				Expect(startByte).To(Equal(int64(10)))
				Expect(endByte).To(Equal(int64(15)))
			})
		})
	})
})

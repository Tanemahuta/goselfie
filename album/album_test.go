package album

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	testfixture "github.com/tanemahuta/goselfie/internal/test"
	"github.com/tanemahuta/goselfie/snapshot"
)

var _ = Describe("Album", func() {
	Describe("reading snapshot albums", func() {
		When("the last snapshot is truncated", func() {
			var decoded []snapshot.Data
			var decodeErr error

			BeforeEach(func() {
				decoded, decodeErr = Decode(bytes.NewReader(testfixture.ReadTestFile("truncated/album.ss")))
			})

			It("reports the incomplete record rather than accepting end of file", func() {
				Expect(decoded).To(BeNil())
				Expect(decodeErr).To(MatchError(ContainSubstring("snapshot separator")))
			})
		})
	})

	Describe("metadata and empty writes", func() {
		var testFile string
		var opened Album

		BeforeEach(func() {
			testFile = filepath.Join(GinkgoT().TempDir(), "example_test.go")
			var err error
			opened, err = OpenAlbum(testFile)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() { Expect(opened.Close()).To(Succeed()) })

		It("exposes the test and archive paths without creating an empty archive", func() {
			Expect(opened.TestFilePath()).To(Equal(testFile))
			Expect(opened.FilePath()).To(Equal(filepath.Join(filepath.Dir(testFile), SnapshotsDirName, "example_test.ss")))
			Expect(opened.Write()).To(Succeed())
			Expect(opened.FilePath()).NotTo(BeAnExistingFile())
		})
	})

	Describe("opening stored snapshots", func() {
		When("the album file contains multiple snapshots", func() {
			var testFile string
			var loaded Album

			BeforeEach(func() {
				testFile = filepath.Join(GinkgoT().TempDir(), "example_test.go")
				albumFile := filepath.Join(filepath.Dir(testFile), SnapshotsDirName, "example_test.ss")
				Expect(os.MkdirAll(filepath.Dir(albumFile), 0o755)).To(Succeed())

				var encoded bytes.Buffer
				Expect(Encode([]snapshot.Data{
					testData{name: []string{"suite", "first"}, contents: testfixture.ReadTestFile("stored/text.txt")},
					testData{name: []string{"suite", "second"}, contents: testfixture.ReadTestFile("stored/binary.bin")},
				}, &encoded)).To(Succeed())
				Expect(os.WriteFile(albumFile, encoded.Bytes(), 0o600)).To(Succeed())

				var err error
				loaded, err = OpenAlbum(testFile)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() { Expect(loaded.Close()).To(Succeed()) })

			It("exposes each stored snapshot", func() {
				expected := []testData{
					{name: []string{"suite", "first"}, contents: testfixture.ReadTestFile("stored/text.txt")},
					{name: []string{"suite", "second"}, contents: testfixture.ReadTestFile("stored/binary.bin")},
				}
				for _, value := range expected {
					actual := loaded.SnapshotByName(value.Name())
					Expect(actual).NotTo(BeNil())
					Expect(actual.Contents()).To(Equal(value.contents))
				}
			})
		})
	})

	Describe("writing updated snapshots", func() {
		var albumFile string
		var opened Album
		var storedFirst, storedSecond snapshot.Data
		var updatedFirst, addedThird snapshot.Taken

		BeforeEach(func() {
			testFile := filepath.Join(GinkgoT().TempDir(), "example_test.go")
			albumFile = filepath.Join(filepath.Dir(testFile), SnapshotsDirName, "example_test.ss")
			Expect(os.MkdirAll(filepath.Dir(albumFile), 0o755)).To(Succeed())

			var initial bytes.Buffer
			Expect(Encode([]snapshot.Data{
				testData{name: []string{"suite", "first It", "#1"}, contents: testfixture.ReadTestFile("updates/old.txt")},
				testData{name: []string{"suite", "first It", "#3"}, contents: testfixture.ReadTestFile("updates/stale.txt")},
				testData{name: []string{"suite", "second It", "#1"}, contents: testfixture.ReadTestFile("updates/preserved.txt")},
			}, &initial)).To(Succeed())
			Expect(os.WriteFile(albumFile, initial.Bytes(), 0o600)).To(Succeed())

			var err error
			opened, err = OpenAlbum(testFile)
			Expect(err).NotTo(HaveOccurred())
			storedFirst = opened.SnapshotByName([]string{"suite", "first It", "#1"})
			storedSecond = opened.SnapshotByName([]string{"suite", "second It", "#1"})
			updatedFirst = newTaken([]string{"suite", "first It", "#1"}, string(testfixture.ReadTestFile("updates/new.txt")))
			addedThird = newTaken([]string{"suite", "first It", "#2"}, string(testfixture.ReadTestFile("updates/added.txt")))
			Expect(opened.UpdateSnapshot(updatedFirst)).To(Succeed())
			Expect(opened.UpdateSnapshot(addedThird)).To(Succeed())
		})

		AfterEach(func() { Expect(opened.Close()).To(Succeed()) })

		When("the album is written and closed", func() {
			BeforeEach(func() {
				Expect(opened.Write()).To(Succeed())
				Expect(opened.Close()).To(Succeed())
			})

			It("stores the updates and snapshots from other containers", func() {
				file, err := os.Open(albumFile)
				Expect(err).NotTo(HaveOccurred())
				decoded, err := Decode(file)
				Expect(file.Close()).To(Succeed())
				Expect(err).NotTo(HaveOccurred())
				DeferCleanup(func() { evictSnapshots(decoded) })

				decodedByName := make(map[string]snapshot.Data, len(decoded))
				for _, current := range decoded {
					decodedByName[strings.Join(current.Name(), "\x00")] = current
				}
				Expect(decodedByName).To(HaveLen(3))
				Expect(decodedByName["suite\x00first It\x00#1"].Contents()).To(Equal(testfixture.ReadTestFile("updates/new.txt")))
				Expect(decodedByName["suite\x00first It\x00#2"].Contents()).To(Equal(testfixture.ReadTestFile("updates/added.txt")))
				Expect(decodedByName["suite\x00second It\x00#1"].Contents()).To(Equal(testfixture.ReadTestFile("updates/preserved.txt")))
				Expect(decodedByName).NotTo(HaveKey("suite\x00first It\x00#3"))
			})

			It("evicts all temporary snapshots", func() {
				for _, data := range []snapshot.Data{storedFirst, storedSecond, updatedFirst, addedThird} {
					_, err := data.Contents()
					Expect(err).To(HaveOccurred())
				}
			})
		})
	})

	Describe("closing an unchanged album", func() {
		var albumFile string
		var contents []byte
		var opened Album
		var stored snapshot.Data

		When("an existing album is opened without updates", func() {
			BeforeEach(func() {
				testFile := filepath.Join(GinkgoT().TempDir(), "example_test.go")
				albumFile = filepath.Join(filepath.Dir(testFile), SnapshotsDirName, "example_test.ss")
				Expect(os.MkdirAll(filepath.Dir(albumFile), 0o755)).To(Succeed())
				contents = testfixture.ReadTestFile("unchanged/album.ss")
				Expect(os.WriteFile(albumFile, contents, 0o600)).To(Succeed())

				var err error
				opened, err = OpenAlbum(testFile)
				Expect(err).NotTo(HaveOccurred())
				stored = opened.SnapshotByName([]string{"value"})
			})

			AfterEach(func() { Expect(opened.Close()).To(Succeed()) })

			When("Write is called without updates", func() {
				It("leaves the existing archive unchanged", func() {
					Expect(opened.Write()).To(Succeed())
					Expect(os.ReadFile(albumFile)).To(Equal(contents))
				})
			})

			When("the album is closed", func() {
				BeforeEach(func() { Expect(opened.Close()).To(Succeed()) })

				It("evicts the temporary snapshot", func() {
					_, err := stored.Contents()
					Expect(err).To(HaveOccurred())
				})

				It("leaves the album file unchanged", func() {
					Expect(os.ReadFile(albumFile)).To(Equal(contents))
				})
			})
		})
	})

	Describe("write failures", func() {
		When("an updated snapshot cannot be encoded", func() {
			var albumFile string
			var original []byte
			var opened *album

			BeforeEach(func() {
				albumFile = filepath.Join(GinkgoT().TempDir(), SnapshotsDirName, "example_test.ss")
				Expect(os.MkdirAll(filepath.Dir(albumFile), 0o755)).To(Succeed())
				original = testfixture.ReadTestFile("write_failure/album.ss")
				Expect(os.WriteFile(albumFile, original, 0o600)).To(Succeed())
				opened = &album{
					filePath:  albumFile,
					snapshots: map[string]snapshot.Data{},
					updated: map[string]snapshot.Taken{
						"broken": failingTaken{testData: testData{name: []string{"broken"}, err: errors.New("broken contents")}},
					},
				}
			})

			It("preserves the existing album file", func() {
				writeErr := opened.Write()
				Expect(writeErr).To(MatchError(ContainSubstring("broken contents")))
				Expect(os.ReadFile(albumFile)).To(Equal(original))
			})
		})
	})

	Describe("listing updated snapshots", func() {
		When("snapshots were updated in reverse name order", func() {
			var opened *album

			BeforeEach(func() {
				opened = &album{updated: map[string]snapshot.Taken{}}
				opened.updated[opened.nameToKey([]string{"second"})] = newTaken([]string{"second"}, "2")
				opened.updated[opened.nameToKey([]string{"first"})] = newTaken([]string{"first"}, "1")
			})
			AfterEach(func() { Expect(opened.Close()).To(Succeed()) })

			It("returns snapshots in deterministic name order", func() {
				updated := opened.UpdatedSnapshots()
				Expect(updated).To(HaveLen(2))
				Expect(updated[0].Name()).To(Equal([]string{"first"}))
				Expect(updated[1].Name()).To(Equal([]string{"second"}))
			})
		})
	})
})

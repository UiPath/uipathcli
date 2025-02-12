package plugin

import "os"

type Archive interface {
	Extract(filePath string, destinationFolder string, permissions os.FileMode) error
}

func newArchive(archiveType ArchiveType) Archive {
	if archiveType == ArchiveTypeZip {
		return newZipArchive()
	}
	return newTarGzArchive()
}

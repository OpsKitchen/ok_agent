package returndata

type File struct {
	FilePath    string
	User        string
	UserGroup   string
	Mode        string
	FileType    string
	FileContent string
	NoTruncate  bool
	Target      string
}

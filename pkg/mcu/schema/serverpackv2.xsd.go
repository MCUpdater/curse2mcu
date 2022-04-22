package schema

// ServerPack ...
type ServerPack struct {
	VersionAttr string        `xml:"version,attr"`
	Server      []*ServerType `xml:"Server"`
}

// Import ...
type Import struct {
	UrlAttr string `xml:"url,attr,omitempty"`
	Value   string `xml:",chardata"`
}

// ServerType ...
type ServerType struct {
	IdAttr            string        `xml:"id,attr"`
	AbstractAttr      bool          `xml:"abstract,attr,omitempty"`
	NameAttr          string        `xml:"name,attr"`
	NewsUrlAttr       string        `xml:"newsUrl,attr,omitempty"`
	IconUrlAttr       string        `xml:"iconUrl,attr,omitempty"`
	VersionAttr       string        `xml:"version,attr"`
	ServerAddressAttr string        `xml:"serverAddress,attr,omitempty"`
	GenerateListAttr  bool          `xml:"generateList,attr,omitempty"`
	AutoConnectAttr   bool          `xml:"autoConnect,attr,omitempty"`
	RevisionAttr      string        `xml:"revision,attr"`
	MainClassAttr     string        `xml:"mainClass,attr,omitempty"`
	LauncherTypeAttr  string        `xml:"launcherType,attr,omitempty"`
	LibOverridesAttr  string        `xml:"libOverrides,attr,omitempty"`
	ServerClassAttr   string        `xml:"serverClass,attr,omitempty"`
	Loader            []*LoaderType `xml:"Loader"`
	Import            []*Import     `xml:"Import"`
	Module            []*ModuleType `xml:"Module"`
}

// LoaderType ...
type LoaderType struct {
	TypeAttr      string `xml:"type,attr"`
	VersionAttr   string `xml:"version,attr"`
	LoadOrderAttr int    `xml:"loadOrder,attr"`
}

// ModuleType ...
type ModuleType struct {
	Submodule  []*ModuleGenericType `xml:"Submodule"`
	ConfigFile []*ConfigType        `xml:"ConfigFile"`
	*ModuleGenericType
}

// URL ...
type URL struct {
	PriorityAttr int    `xml:"priority,attr,omitempty"`
	Value        string `xml:",chardata"`
}

// Curse ...
type Curse struct {
	ProjectAttr     string `xml:"project,attr,omitempty"`
	FileAttr        int    `xml:"file,attr,omitempty"`
	TypeAttr        string `xml:"type,attr,omitempty"`
	AutoupgradeAttr bool   `xml:"autoupgrade,attr,omitempty"`
}

// Required ...
type Required struct {
	IsDefaultAttr bool `xml:"isDefault,attr,omitempty"`
	Value         bool `xml:",chardata"`
}

// ModType ...
type ModType struct {
	InRootAttr     bool   `xml:"inRoot,attr,omitempty"`
	OrderAttr      int    `xml:"order,attr,omitempty"`
	KeepMetaAttr   bool   `xml:"keepMeta,attr,omitempty"`
	LaunchArgsAttr string `xml:"launchArgs,attr,omitempty"`
	JreArgsAttr    string `xml:"jreArgs,attr,omitempty"`
	*ModEnum       `xml:",chardata"`
}

// ModuleGenericType ...
type ModuleGenericType struct {
	NameAttr    string    `xml:"name,attr"`
	IdAttr      string    `xml:"id,attr"`
	DependsAttr string    `xml:"depends,attr,omitempty"`
	SideAttr    string    `xml:"side,attr,omitempty"`
	URL         []*URL    `xml:"URL"`
	Curse       *Curse    `xml:"Curse"`
	LoadPrefix  string    `xml:"LoadPrefix,omitempty"`
	ModPath     string    `xml:"ModPath,omitempty"`
	Size        int64     `xml:"Size"`
	Required    *Required `xml:"Required"`
	ModType     *ModType  `xml:"ModType"`
	MD5         string    `xml:"MD5"`
	Meta        *MetaType `xml:"Meta"`
}

// ConfigType ...
type ConfigType struct {
	URL         []*URL `xml:"URL"`
	Path        string `xml:"Path"`
	NoOverwrite bool   `xml:"NoOverwrite"`
	MD5         string `xml:"MD5"`
}

// MetaType ...
type MetaType struct {
}

// ModEnum ...
type ModEnum string

var (
	ModTypeRegular = ModEnum("Regular")
	ModTypeExtract = ModEnum("Extract")
)

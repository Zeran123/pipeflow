package bot

const dbPath = "data.db"
const dbBucket = "Bots"

type Store struct {
	Source string
	Target string
	Url    string
	Id     string
}

var db Db = Blot{dbPath, dbBucket}

func (s Store) Save() (Store, error) {
	return db.save(s)
}

func List() ([]Store, error) {
	return db.list()
}

func (s Store) Del() error {
	return db.del(s.Id)
}

func Get(id string) (Store, error) {
	return db.get(id)
}

package je

type RemoteStore struct {
	client *Client
}

func (store *RemoteStore) Close() error {
	return nil
}

func (store *RemoteStore) NextId() ID {
	return 0
}

func (store *RemoteStore) Save(job *Job) error {
	_, err := store.client.UpdateJob(job)
	return err
}

func (store *RemoteStore) Get(id ID) (job *Job, err error) {
	jobs, err := store.client.GetJobByID(string(id))
	if len(jobs) > 0 && err == nil {
		return jobs[0], nil
	}
	return
}

func (store *RemoteStore) Find(ids ...ID) (jobs []*Job, err error) {
	return
}

func (store *RemoteStore) All() (jobs []*Job, err error) {
	return
}

func (store *RemoteStore) Search(q string) (jobs []*Job, err error) {
	return
}

func NewRemoteStore(uri string) (Store, error) {
	return &RemoteStore{
		client: NewClient(uri, nil),
	}, nil
}

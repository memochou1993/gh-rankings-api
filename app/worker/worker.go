package worker

var (
	Repository *RepositoryWorker
	Owner      *OwnerWorker
)

func Init() {
	Owner = NewOwnerWorker()
	Owner.Init()
	Owner.Work()
	Repository = NewRepositoryWorker()
	Repository.Init()
	Repository.Work()
}

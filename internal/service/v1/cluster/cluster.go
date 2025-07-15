package cluster

import (
	v1Cluster "github.com/KubeOperator/kubepi/internal/model/v1/cluster"
	"github.com/KubeOperator/kubepi/internal/service/v1/common"
	"github.com/KubeOperator/kubepi/pkg/storm"
	"github.com/KubeOperator/kubepi/pkg/util/lang"
	"github.com/asdine/storm/v3/q"
	"github.com/google/uuid"
	"time"
)

type Service interface {
	common.DBService
	Create(cluster *v1Cluster.Cluster, options common.DBOptions) error
	Update(name string, cluster *v1Cluster.Cluster, options common.DBOptions) error
	Get(name string, options common.DBOptions) (*v1Cluster.Cluster, error)
	List(options common.DBOptions) ([]v1Cluster.Cluster, error)
	Delete(name string, options common.DBOptions) error
	Search(num, size int, conditions common.Conditions, options common.DBOptions) ([]v1Cluster.Cluster, int, error)
}

func NewService() Service {
	return &cluster{
		DefaultDBService: common.DefaultDBService{},
	}
}

type cluster struct {
	common.DefaultDBService
}

func (c *cluster) Update(name string, cluster *v1Cluster.Cluster, options common.DBOptions) error {
	db := c.GetDB(options)
	r, err := c.Get(name, options)
	if err != nil {
		return err
	}
	cluster.UUID = r.UUID
	cluster.CreateAt = r.CreateAt
	cluster.UpdateAt = time.Now()
	return db.Update(cluster)
}

func (c *cluster) Create(cluster *v1Cluster.Cluster, options common.DBOptions) error {
	db := c.GetDB(options)
	cluster.UUID = uuid.New().String()
	cluster.CreateAt = time.Now()
	cluster.UpdateAt = time.Now()
	return db.Save(cluster)
}

func (c *cluster) Get(name string, options common.DBOptions) (*v1Cluster.Cluster, error) {
	db := c.GetDB(options)
	var cluster v1Cluster.Cluster
	if err := db.One("Name", name, &cluster); err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (c *cluster) List(options common.DBOptions) ([]v1Cluster.Cluster, error) {
	db := c.GetDB(options)
	var clusters []v1Cluster.Cluster
	if err := db.All(&clusters); err != nil {
		return nil, err
	}
	return clusters, nil
}

func (c *cluster) Search(num, size int, conditions common.Conditions, options common.DBOptions) ([]v1Cluster.Cluster, int, error) {
	db := c.GetDB(options)

	var ms []q.Matcher
	
	// Check the type of conditions and handle accordingly
	switch cond := conditions.(type) {
	case common.ConditionsMap:
		// Handle map-based conditions (old format)
		for k := range cond {
			if k == "quick" {
				ms = append(ms, storm.Like("Name", cond[k].Value))
			} else if k == "labels" {
				switch cond[k].Operator {
				case "like":
					ms = append(ms, storm.ArrayValueLike("Labels", cond[k].Value))
				case "not like":
					ms = append(ms, q.Not(storm.ArrayValueLike("Labels", cond[k].Value)))
				case "eq":
					ms = append(ms, storm.ArrayValueEq("Labels", cond[k].Value))
				case "ne":
					ms = append(ms, q.Not(storm.ArrayValueEq("Labels", cond[k].Value)))
				}
			} else {
				field := lang.FirstToUpper(cond[k].Field)
				switch cond[k].Operator {
				case "eq":
					ms = append(ms, q.Eq(field, cond[k].Value))
				case "ne":
					ms = append(ms, q.Not(q.Eq(field, cond[k].Value)))
				case "like":
					ms = append(ms, storm.Like(field, cond[k].Value))
				case "not like":
					ms = append(ms, q.Not(storm.Like(field, cond[k].Value)))
				}
			}
		}
	case []common.Condition:
		// Handle slice-based conditions (new format)
		for _, condition := range cond {
			if condition.Field == "quick" {
				ms = append(ms, storm.Like("Name", condition.Value))
			} else if condition.Field == "labels" {
				switch condition.Operator {
				case "like":
					ms = append(ms, storm.ArrayValueLike("Labels", condition.Value))
				case "not like":
					ms = append(ms, q.Not(storm.ArrayValueLike("Labels", condition.Value)))
				case "eq":
					ms = append(ms, storm.ArrayValueEq("Labels", condition.Value))
				case "ne":
					ms = append(ms, q.Not(storm.ArrayValueEq("Labels", condition.Value)))
				}
			} else {
				field := lang.FirstToUpper(condition.Field)
				switch condition.Operator {
				case "eq":
					ms = append(ms, q.Eq(field, condition.Value))
				case "ne":
					ms = append(ms, q.Not(q.Eq(field, condition.Value)))
				case "like":
					ms = append(ms, storm.Like(field, condition.Value))
				case "not like":
					ms = append(ms, q.Not(storm.Like(field, condition.Value)))
				}
			}
		}
	default:
		// If no conditions are provided, return empty matcher
	}
	
	query := db.Select(ms...).OrderBy("CreateAt").Reverse()
	count, err := query.Count(&v1Cluster.Cluster{})
	if err != nil {
		return nil, 0, err
	}
	if size != 0 {
		query.Limit(size).Skip((num - 1) * size)
	}
	clusters := make([]v1Cluster.Cluster, 0)
	if err := query.Find(&clusters); err != nil {
		return clusters, 0, err
	}
	return clusters, count, nil
}

func (c *cluster) Delete(name string, options common.DBOptions) error {
	db := c.GetDB(options)
	cluster, err := c.Get(name, options)
	if err != nil {
		return err
	}
	return db.DeleteStruct(cluster)
}

package domain

import "sort"

type Doc struct {
	ID          string   `json:"id" reindex:"id,,pk"`
	Title       string   `json:"title" reindex:"title"`
	Sort        int32    `json:"sort,omitempty" reindex:"sort"`
	ChildrenIDs []string `json:"children_ids,omitempty" reindex:"children_ids"`
	SubDocs     []*Doc   `json:"sub_docs,omitempty" reindex:"sub_docs,,joined"`
}

func (d *Doc) Finalize() {
	d.Sort = 0
	d.ChildrenIDs = nil
	sort.Slice(d.SubDocs, func(i, j int) bool {
		return d.SubDocs[i].Sort > d.SubDocs[j].Sort
	})

	for _, subDoc := range d.SubDocs {
		subDoc.ChildrenIDs = nil
		for _, subSubDoc := range subDoc.SubDocs {
			subSubDoc.Sort = 0
			subDoc.ChildrenIDs = nil
		}
	}
}

package main

import (
	"time"
)

// The Annotation model

type RelationshipNameKey struct {
	Index            *int   `json:"index" bson:"index"`
	RelationshipName string `json:"relationship_name" bson:"relationship_name"`
}

type ClusterRelationship struct {
	TokenClusters     []*int                `json:"token_clusters" bson:"token_clusters"`
	RelationshipNames []RelationshipNameKey `json:"relationship_names" bson:"relationship_names"`
	Index             *int                  `json:"index" bson:"index"`
}

type TokenCluster struct {
	Tokens                  []*int `json:"tokens" bson:"tokens"`
	Name                    string `json:"name" bson:"name"`
	Tore                    string `json:"tore" bson:"tore"`
	Index                   *int   `json:"index" bson:"index"`
	RelationshipMemberships []*int `json:"relationship_memberships" bson:"relationship_memberships"`
}

type Token struct {
	Index        *int   `json:"index" bson:"index"`
	Name         string `validate:"nonzero" json:"name" bson:"name"`
	Lemma        string `validate:"nonzero" json:"lemma" bson:"lemma"`
	Pos          string `validate:"nonzero" json:"pos" bson:"pos"`
	TokenCluster *int   `json:"token_cluster" bson:"token_cluster"`
}

type Annotation struct {
	UploadedAt time.Time `validate:"nonzero" json:"uploaded_at" bson:"uploaded_at"`
	Name       string    `validate:"nonzero" json:"name" bson:"name"`

	Tokens               []Token               `json:"tokens" bson:"tokens"`
	TokenClusters        []TokenCluster        `json:"token_clusters" bson:"token_clusters"`
	ClusterRelationships []ClusterRelationship `json:"cluster_relationships" bson:"cluster_relationships"`
}

// end Annotation model

// Dataset model
type Dataset struct {
	UploadedAt  time.Time      `json:"uploaded_at"`
	Name        string         `json:"name"`
	Size        int            `json:"size"`
	Documents   []Document     `json:"documents"`
	GroundTruth []TruthElement `json:"ground_truth" bson:"ground_truth"`
}

//TruthElement model
type TruthElement struct {
	Id    string `json:"id" bson:"id"`
	Value string `json:"value"  bson:"value"`
}

// Document model
type Document struct {
	Number int    `json:"number"`
	Text   string `json:"text"`
	Id     string `json:"id"`
}

// Result model
type Result struct {
	Method      string                 `json:"method"`
	Status      string                 `json:"status"`
	StartedAt   time.Time              `json:"started_at"`
	DatasetName string                 `json:"dataset_name"`
	Params      map[string]string      `json:"params"`
	Topics      map[string]interface{} `json:"topics"`
	DocTopic    map[string]interface{} `json:"doc_topic"`
	Metrics     map[string]interface{} `json:"metrics"`
	Name        string                 `json:"name"`
}

// Run model
type Run struct {
	Method  string            `json:"method"`
	Dataset Dataset           `json:"dataset"`
	Params  map[string]string `json:"params"`
}

// ResponseMessage model
type ResponseMessage struct {
	Message string `json:"message"`
	Status  bool   `json:"status"`
}

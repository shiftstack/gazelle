package main

import "time"

type Run struct {
	JobName       string    `json:"jobName"`
	RunNumber     string    `json:"runNumber"`
	Started       time.Time `json:"started"`
	Duration      string    `json:"duration"`
	Result        string    `json:"result"`
	BuildLogURL   string    `json:"buildLogURL"`
	MatchingRules []string  `json:"matchingRules"`
}

func (r *Run) Rules() {}

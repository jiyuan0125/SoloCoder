package toposort

import (
	"errors"
	"fmt"
	"kahn-toposort/pkg/common"
)

type TopoSorter struct {
	tasks        map[string]bool
	dependencies map[string]map[string]bool
	inDegree     map[string]int
	adjList      map[string][]string
}

func NewTopoSorter() *TopoSorter {
	return &TopoSorter{
		tasks:        make(map[string]bool),
		dependencies: make(map[string]map[string]bool),
		inDegree:     make(map[string]int),
		adjList:      make(map[string][]string),
	}
}

func (ts *TopoSorter) Import(tasks []string, deps []common.Dependency) error {
	ts.tasks = make(map[string]bool)
	ts.dependencies = make(map[string]map[string]bool)
	ts.inDegree = make(map[string]int)
	ts.adjList = make(map[string][]string)

	for _, task := range tasks {
		ts.tasks[task] = true
		ts.inDegree[task] = 0
		ts.adjList[task] = []string{}
	}

	for _, dep := range deps {
		if !ts.tasks[dep.Before] {
			return errors.New(fmt.Sprintf("任务不存在: %s (依赖项before引用了不存在的任务)", dep.Before))
		}
		if !ts.tasks[dep.After] {
			return errors.New(fmt.Sprintf("任务不存在: %s (依赖项after引用了不存在的任务)", dep.After))
		}

		if dep.Before == dep.After {
			return errors.New(fmt.Sprintf("循环依赖: %s 不能依赖自身", dep.Before))
		}

		depKey := fmt.Sprintf("%s->%s", dep.Before, dep.After)
		if ts.dependencies[depKey] == nil {
			ts.dependencies[depKey] = make(map[string]bool)
		}
		if ts.dependencies[depKey]["exists"] {
			continue
		}
		ts.dependencies[depKey]["exists"] = true

		ts.inDegree[dep.After]++
		ts.adjList[dep.Before] = append(ts.adjList[dep.Before], dep.After)
	}

	return nil
}

func (ts *TopoSorter) Sort() ([][]string, bool) {
	levels := [][]string{}
	queue := []string{}

	for task, degree := range ts.inDegree {
		if degree == 0 {
			queue = append(queue, task)
		}
	}

	totalTasks := len(ts.tasks)
	processed := 0

	for len(queue) > 0 {
		currentLevelSize := len(queue)
		currentLevel := []string{}

		for i := 0; i < currentLevelSize; i++ {
			node := queue[0]
			queue = queue[1:]
			currentLevel = append(currentLevel, node)
			processed++

			for _, neighbor := range ts.adjList[node] {
				ts.inDegree[neighbor]--
				if ts.inDegree[neighbor] == 0 {
					queue = append(queue, neighbor)
				}
			}
		}

		levels = append(levels, currentLevel)
	}

	hasCycle := processed < totalTasks
	return levels, hasCycle
}

func (ts *TopoSorter) FindCycle() []string {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for task := range ts.tasks {
		if !visited[task] {
			if path := ts.dfsCycle(task, visited, recStack); path != nil {
				return path
			}
		}
	}

	return nil
}

func (ts *TopoSorter) dfsCycle(node string, visited, recStack map[string]bool) []string {
	visited[node] = true
	recStack[node] = true

	for _, neighbor := range ts.adjList[node] {
		if !visited[neighbor] {
			if path := ts.dfsCycle(neighbor, visited, recStack); path != nil {
				return append([]string{node}, path...)
			}
		} else if recStack[neighbor] {
			return []string{node, neighbor}
		}
	}

	recStack[node] = false
	return nil
}

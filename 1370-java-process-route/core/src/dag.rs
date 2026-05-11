use crate::error::{ProcessRouteError, Result};
use crate::models::{CriticalPathInfo, ProcessDefinition};
use petgraph::algo::toposort;
use petgraph::graph::{DiGraph, NodeIndex};
use std::collections::{HashMap, HashSet};

pub struct ProcessDag {
    graph: DiGraph<ProcessNode, ()>,
    id_to_index: HashMap<String, NodeIndex>,
    index_to_id: HashMap<NodeIndex, String>,
}

struct ProcessNode {
    id: String,
    name: String,
    duration: u32,
}

impl ProcessDag {
    pub fn new(processes: &[ProcessDefinition]) -> Result<Self> {
        let mut graph = DiGraph::new();
        let mut id_to_index = HashMap::new();
        let mut index_to_id = HashMap::new();

        for process in processes {
            let node = ProcessNode {
                id: process.id.clone(),
                name: process.name.clone(),
                duration: process.standard_time_minutes,
            };
            let idx = graph.add_node(node);
            id_to_index.insert(process.id.clone(), idx);
            index_to_id.insert(idx, process.id.clone());
        }

        for process in processes {
            let current_idx = id_to_index
                .get(&process.id)
                .ok_or_else(|| ProcessRouteError::ProcessNotFound(process.id.clone()))?;

            for pred_id in &process.predecessors {
                let pred_idx = id_to_index
                    .get(pred_id)
                    .ok_or_else(|| ProcessRouteError::ProcessNotFound(pred_id.clone()))?;
                graph.add_edge(*pred_idx, *current_idx, ());
            }
        }

        if toposort(&graph, None).is_err() {
            return Err(ProcessRouteError::CircularDependency);
        }

        Ok(Self {
            graph,
            id_to_index,
            index_to_id,
        })
    }

    pub fn calculate_critical_path(&self) -> Result<CriticalPathInfo> {
        let topo_order = toposort(&self.graph, None)
            .map_err(|_| ProcessRouteError::CircularDependency)?;

        let mut earliest_finish: HashMap<NodeIndex, u32> = HashMap::new();
        let mut predecessors_for_node: HashMap<NodeIndex, HashSet<NodeIndex>> = HashMap::new();

        for &node in &topo_order {
            let node_data = &self.graph[node];
            let duration = node_data.duration;

            let max_pred_finish = self
                .graph
                .neighbors_directed(node, petgraph::Direction::Incoming)
                .map(|pred| earliest_finish.get(&pred).copied().unwrap_or(0))
                .max()
                .unwrap_or(0);

            earliest_finish.insert(node, max_pred_finish + duration);

            self.graph
                .neighbors_directed(node, petgraph::Direction::Incoming)
                .for_each(|pred| {
                    predecessors_for_node
                        .entry(node)
                        .or_insert_with(HashSet::new)
                        .insert(pred);
                });
        }

        let end_node = topo_order
            .iter()
            .max_by_key(|&&n| earliest_finish.get(&n).copied().unwrap_or(0))
            .copied()
            .ok_or_else(|| ProcessRouteError::Internal("工艺路线为空".to_string()))?;

        let total_time = *earliest_finish
            .get(&end_node)
            .ok_or_else(|| ProcessRouteError::Internal("无法计算总时间".to_string()))?;

        let mut critical_path = Vec::new();
        let mut current_node = end_node;
        let mut visited = HashSet::new();

        loop {
            if visited.contains(&current_node) {
                break;
            }
            visited.insert(current_node);

            let node_id = self
                .index_to_id
                .get(&current_node)
                .ok_or_else(|| ProcessRouteError::Internal("节点映射错误".to_string()))?
                .clone();
            critical_path.push(node_id);

            let current_finish = earliest_finish.get(&current_node).copied().unwrap_or(0);
            let current_duration = self.graph[current_node].duration;
            let needed_pred_finish = current_finish - current_duration;

            let preds = predecessors_for_node.get(&current_node);
            if preds.is_none() || preds.unwrap().is_empty() {
                break;
            }

            let preds = preds.unwrap();
            let critical_pred = preds
                .iter()
                .find(|&&pred| {
                    earliest_finish.get(&pred).copied().unwrap_or(0) == needed_pred_finish
                })
                .copied();

            match critical_pred {
                Some(pred) => current_node = pred,
                None => break,
            }
        }

        critical_path.reverse();

        Ok(CriticalPathInfo {
            path: critical_path,
            total_time_minutes: total_time,
        })
    }

    pub fn get_critical_process_ids(&self) -> Result<HashSet<String>> {
        let info = self.calculate_critical_path()?;
        Ok(info.path.into_iter().collect())
    }

    pub fn get_start_processes(&self) -> Vec<String> {
        self.graph
            .node_indices()
            .filter(|&n| {
                self.graph
                    .neighbors_directed(n, petgraph::Direction::Incoming)
                    .count()
                    == 0
            })
            .filter_map(|n| self.index_to_id.get(&n).cloned())
            .collect()
    }

    pub fn get_end_processes(&self) -> Vec<String> {
        self.graph
            .node_indices()
            .filter(|&n| {
                self.graph
                    .neighbors_directed(n, petgraph::Direction::Outgoing)
                    .count()
                    == 0
            })
            .filter_map(|n| self.index_to_id.get(&n).cloned())
            .collect()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_simple_dag_critical_path() {
        let processes = vec![
            ProcessDefinition::with_id("A", "工序A", 5, vec![]),
            ProcessDefinition::with_id("B", "工序B", 3, vec!["A".to_string()]),
            ProcessDefinition::with_id("C", "工序C", 4, vec!["A".to_string()]),
            ProcessDefinition::with_id("D", "工序D", 2, vec!["B".to_string(), "C".to_string()]),
        ];

        let dag = ProcessDag::new(&processes).unwrap();
        let critical = dag.calculate_critical_path().unwrap();

        assert_eq!(critical.total_time_minutes, 11);
        assert_eq!(critical.path, vec!["A", "C", "D"]);
    }

    #[test]
    fn test_circular_dependency() {
        let processes = vec![
            ProcessDefinition::with_id("A", "工序A", 5, vec!["B".to_string()]),
            ProcessDefinition::with_id("B", "工序B", 3, vec!["A".to_string()]),
        ];

        let result = ProcessDag::new(&processes);
        assert!(matches!(result, Err(ProcessRouteError::CircularDependency)));
    }
}

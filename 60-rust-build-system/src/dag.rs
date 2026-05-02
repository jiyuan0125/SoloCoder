use crate::config::{BuildConfig, TargetConfig};
use std::collections::{HashMap, HashSet, VecDeque};

#[derive(Debug, Clone)]
pub struct Node {
    pub name: String,
    pub config: TargetConfig,
    pub dependencies: Vec<String>,
    pub dependents: Vec<String>,
}

#[derive(Debug)]
pub struct BuildDAG {
    pub nodes: HashMap<String, Node>,
    pub output_to_target: HashMap<String, String>,
}

#[derive(Debug, Clone)]
pub struct CycleError {
    pub cycle: Vec<String>,
}

impl std::fmt::Display for CycleError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "Cycle detected: ")?;
        for (i, name) in self.cycle.iter().enumerate() {
            if i > 0 {
                write!(f, " -> ")?;
            }
            write!(f, "{}", name)?;
        }
        Ok(())
    }
}

impl BuildDAG {
    pub fn from_config(config: &BuildConfig) -> Result<Self, String> {
        let mut nodes = HashMap::new();
        let mut output_to_target = HashMap::new();

        for target in &config.targets {
            for output in &target.outputs {
                if let Some(existing) = output_to_target.insert(output.clone(), target.name.clone()) {
                    return Err(format!(
                        "Output file '{}' is produced by both '{}' and '{}'",
                        output, existing, target.name
                    ));
                }
            }

            nodes.insert(
                target.name.clone(),
                Node {
                    name: target.name.clone(),
                    config: target.clone(),
                    dependencies: Vec::new(),
                    dependents: Vec::new(),
                },
            );
        }

        let mut dag = BuildDAG {
            nodes,
            output_to_target,
        };

        dag.derive_dependencies()?;
        if let Err(cycle_err) = dag.check_for_cycles() {
            return Err(cycle_err.to_string());
        }

        Ok(dag)
    }

    fn derive_dependencies(&mut self) -> Result<(), String> {
        let target_names: Vec<String> = self.nodes.keys().cloned().collect();
        
        for target_name in &target_names {
            let node = self.nodes.get(target_name).unwrap();
            let inputs = &node.config.inputs;
            
            let mut deps = HashSet::new();
            
            for input in inputs {
                if let Some(producer) = self.output_to_target.get(input) {
                    if producer != target_name {
                        deps.insert(producer.clone());
                    }
                }
            }

            let node = self.nodes.get_mut(target_name).unwrap();
            node.dependencies = deps.iter().cloned().collect();

            for dep in deps {
                let dep_node = self.nodes.get_mut(&dep).unwrap();
                if !dep_node.dependents.contains(target_name) {
                    dep_node.dependents.push(target_name.clone());
                }
            }
        }

        Ok(())
    }

    fn check_for_cycles(&self) -> Result<(), CycleError> {
        let mut visited = HashSet::new();
        let mut rec_stack = HashSet::new();
        let mut path = Vec::new();

        for node_name in self.nodes.keys() {
            if self.dfs_cycle_check(node_name, &mut visited, &mut rec_stack, &mut path) {
                return Err(CycleError { cycle: path.clone() });
            }
        }

        Ok(())
    }

    fn dfs_cycle_check(
        &self,
        node_name: &str,
        visited: &mut HashSet<String>,
        rec_stack: &mut HashSet<String>,
        path: &mut Vec<String>,
    ) -> bool {
        if !visited.contains(node_name) {
            visited.insert(node_name.to_string());
            rec_stack.insert(node_name.to_string());
            path.push(node_name.to_string());

            if let Some(node) = self.nodes.get(node_name) {
                for dep in &node.dependencies {
                    if !visited.contains(dep) && self.dfs_cycle_check(dep, visited, rec_stack, path) {
                        return true;
                    } else if rec_stack.contains(dep) {
                        let idx = path.iter().position(|p| p == dep).unwrap();
                        *path = path[idx..].to_vec();
                        path.push(dep.clone());
                        return true;
                    }
                }
            }
        }

        if rec_stack.contains(node_name) {
            rec_stack.remove(node_name);
            path.pop();
        }

        false
    }

    pub fn topological_order(&self) -> Vec<String> {
        let mut in_degree = HashMap::new();
        for node in self.nodes.values() {
            in_degree.insert(node.name.clone(), node.dependencies.len());
        }

        let mut queue = VecDeque::new();
        for (name, degree) in &in_degree {
            if *degree == 0 {
                queue.push_back(name.clone());
            }
        }

        let mut result = Vec::new();
        while let Some(node_name) = queue.pop_front() {
            result.push(node_name.clone());

            if let Some(node) = self.nodes.get(&node_name) {
                for dependent in &node.dependents {
                    if let Some(degree) = in_degree.get_mut(dependent) {
                        *degree -= 1;
                        if *degree == 0 {
                            queue.push_back(dependent.clone());
                        }
                    }
                }
            }
        }

        result
    }

    pub fn get_ready_nodes(&self, completed: &HashSet<String>) -> Vec<String> {
        let mut ready = Vec::new();
        for node in self.nodes.values() {
            if !completed.contains(&node.name) {
                let all_deps_done = node.dependencies.iter().all(|dep| completed.contains(dep));
                if all_deps_done {
                    ready.push(node.name.clone());
                }
            }
        }
        ready
    }
}

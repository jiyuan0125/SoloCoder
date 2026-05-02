use std::collections::HashMap;

use crate::{Handler, Method, Params, RouterError};
use super::route::{Segment, PathParser};

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum NodeType {
    Static,
    Param,
    Wildcard,
    DoubleWildcard,
}

#[derive(Debug, Clone)]
pub struct TreeNode {
    pub segment: String,
    pub node_type: NodeType,
    pub children: Vec<TreeNode>,
    pub handlers: HashMap<Method, HandlerInfo>,
}

#[derive(Debug, Clone)]
pub struct HandlerInfo {
    pub path: String,
    pub handler: Handler,
    pub handler_name: &'static str,
}

pub struct RadixTree {
    root: TreeNode,
}

impl RadixTree {
    pub fn new() -> Self {
        RadixTree {
            root: TreeNode {
                segment: String::new(),
                node_type: NodeType::Static,
                children: Vec::new(),
                handlers: HashMap::new(),
            },
        }
    }

    pub fn insert(
        &mut self,
        path: &str,
        method: Method,
        handler: Handler,
        handler_name: &'static str,
    ) -> Result<(), RouterError> {
        let segments = PathParser::validate_and_parse(path)?;

        if segments.is_empty() {
            if self.root.handlers.contains_key(&method) {
                return Err(RouterError::RouteAlreadyExists {
                    method,
                    path: path.to_string(),
                });
            }
            self.root.handlers.insert(
                method,
                HandlerInfo {
                    path: path.to_string(),
                    handler,
                    handler_name,
                },
            );
            return Ok(());
        }

        self.insert_segments(&segments, path, method, handler, handler_name)
    }

    fn insert_segments(
        &mut self,
        segments: &[Segment],
        path: &str,
        method: Method,
        handler: Handler,
        handler_name: &'static str,
    ) -> Result<(), RouterError> {
        let mut current = &mut self.root;

        for (i, segment) in segments.iter().enumerate() {
            let (segment_str, node_type) = match segment {
                Segment::Static(s) => (s.clone(), NodeType::Static),
                Segment::Param(name) => (name.clone(), NodeType::Param),
                Segment::Wildcard => ("*".to_string(), NodeType::Wildcard),
                Segment::DoubleWildcard => ("**".to_string(), NodeType::DoubleWildcard),
            };

            let child_index = current
                .children
                .iter()
                .position(|child| child.segment == segment_str && child.node_type == node_type);

            match child_index {
                Some(idx) => {
                    current = &mut current.children[idx];
                }
                None => {
                    let new_node = TreeNode {
                        segment: segment_str,
                        node_type,
                        children: Vec::new(),
                        handlers: HashMap::new(),
                    };
                    current.children.push(new_node);
                    let last_idx = current.children.len() - 1;
                    current = &mut current.children[last_idx];
                }
            }

            if i == segments.len() - 1 {
                if current.handlers.contains_key(&method) {
                    return Err(RouterError::RouteAlreadyExists {
                        method: method.clone(),
                        path: path.to_string(),
                    });
                }
                current.handlers.insert(
                    method.clone(),
                    HandlerInfo {
                        path: path.to_string(),
                        handler,
                        handler_name,
                    },
                );
            }
        }

        Ok(())
    }

    pub fn lookup(&self, path: &str, method: &Method) -> Option<(&HandlerInfo, Params)> {
        let segments: Vec<&str> = path
            .split('/')
            .skip(1)
            .filter(|s| !s.is_empty())
            .collect();

        if segments.is_empty() {
            return self.root.handlers.get(method).map(|hi| (hi, Params::new()));
        }

        self.lookup_segments(&segments, method)
    }

    fn lookup_segments(&self, segments: &[&str], method: &Method) -> Option<(&HandlerInfo, Params)> {
        let mut results: Vec<(&HandlerInfo, Params, u8)> = Vec::new();

        self.search_node(&self.root, segments, 0, Params::new(), method, &mut results);

        if results.is_empty() {
            return None;
        }

        results.sort_by_key(|(_, _, priority)| *priority);
        results.first().map(|(handler, params, _)| (*handler, params.clone()))
    }

    fn search_node<'a>(
        &'a self,
        node: &'a TreeNode,
        segments: &[&str],
        seg_idx: usize,
        params: Params,
        method: &Method,
        results: &mut Vec<(&'a HandlerInfo, Params, u8)>,
    ) {
        if seg_idx >= segments.len() {
            if let Some(handler_info) = node.handlers.get(method) {
                results.push((handler_info, params.clone(), 0));
            }
            return;
        }

        let current_segment = segments[seg_idx];

        for child in &node.children {
            match child.node_type {
                NodeType::Static => {
                    if child.segment == current_segment {
                        self.search_node(child, segments, seg_idx + 1, params.clone(), method, results);
                    }
                }
                NodeType::Param => {
                    let mut new_params = params.clone();
                    new_params.insert(&child.segment, current_segment);
                    self.search_node(child, segments, seg_idx + 1, new_params, method, results);
                }
                NodeType::Wildcard => {
                    if seg_idx == segments.len() - 1 {
                        let mut new_params = params.clone();
                        new_params.insert("*", current_segment);
                        if let Some(handler_info) = child.handlers.get(method) {
                            results.push((handler_info, new_params, 2));
                        }
                    }
                }
                NodeType::DoubleWildcard => {
                    let remaining: Vec<&str> = segments[seg_idx..].to_vec();
                    let wildcard_value = remaining.join("/");
                    let mut new_params = params.clone();
                    new_params.insert("**", wildcard_value);
                    if let Some(handler_info) = child.handlers.get(method) {
                        results.push((handler_info, new_params, 3));
                    }
                }
            }
        }
    }

    pub fn list_routes(&self) -> Vec<RouteEntry> {
        let mut routes = Vec::new();
        self.collect_routes(&self.root, String::new(), &mut routes);
        routes
    }

    fn collect_routes(&self, node: &TreeNode, current_path: String, routes: &mut Vec<RouteEntry>) {
        for (method, handler_info) in &node.handlers {
            let path = if current_path.is_empty() {
                "/".to_string()
            } else {
                current_path.clone()
            };
            routes.push(RouteEntry {
                method: method.clone(),
                path,
                handler_name: handler_info.handler_name,
            });
        }

        for child in &node.children {
            let new_path = if current_path.is_empty() {
                match child.node_type {
                    NodeType::Static => format!("/{}", child.segment),
                    NodeType::Param => format!("/:{}", child.segment),
                    NodeType::Wildcard => "/*".to_string(),
                    NodeType::DoubleWildcard => "/**".to_string(),
                }
            } else {
                match child.node_type {
                    NodeType::Static => format!("{}/{}", current_path, child.segment),
                    NodeType::Param => format!("{}/:{}", current_path, child.segment),
                    NodeType::Wildcard => format!("{}/*", current_path),
                    NodeType::DoubleWildcard => format!("{}/**", current_path),
                }
            };
            self.collect_routes(child, new_path, routes);
        }
    }
}

#[derive(Debug, Clone)]
pub struct RouteEntry {
    pub method: Method,
    pub path: String,
    pub handler_name: &'static str,
}

impl Default for RadixTree {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::{Method, Request, Response};

    fn test_handler(_req: &Request, _params: &Params) -> Response {
        Response::ok("test")
    }

    #[test]
    fn test_insert_and_lookup_static() {
        let mut tree = RadixTree::new();
        tree.insert("/users", Method::GET, test_handler, "test_handler").unwrap();

        let result = tree.lookup("/users", &Method::GET);
        assert!(result.is_some());
        assert_eq!(result.unwrap().1.len(), 0);
    }

    #[test]
    fn test_insert_and_lookup_param() {
        let mut tree = RadixTree::new();
        tree.insert("/users/:id", Method::GET, test_handler, "test_handler").unwrap();

        let result = tree.lookup("/users/42", &Method::GET);
        assert!(result.is_some());
        let (_, params) = result.unwrap();
        assert_eq!(params.get("id"), Some(&"42".to_string()));
    }

    #[test]
    fn test_insert_and_lookup_wildcard() {
        let mut tree = RadixTree::new();
        tree.insert("/files/*", Method::GET, test_handler, "test_handler").unwrap();

        let result = tree.lookup("/files/test.txt", &Method::GET);
        assert!(result.is_some());
        let (_, params) = result.unwrap();
        assert_eq!(params.get("*"), Some(&"test.txt".to_string()));

        let result = tree.lookup("/files/a/b.txt", &Method::GET);
        assert!(result.is_none());
    }

    #[test]
    fn test_insert_and_lookup_double_wildcard() {
        let mut tree = RadixTree::new();
        tree.insert("/static/**", Method::GET, test_handler, "test_handler").unwrap();

        let result = tree.lookup("/static/css/style.css", &Method::GET);
        assert!(result.is_some());
        let (_, params) = result.unwrap();
        assert_eq!(params.get("**"), Some(&"css/style.css".to_string()));

        let result = tree.lookup("/static/js/app.js", &Method::GET);
        assert!(result.is_some());
    }

    #[test]
    fn test_priority() {
        let mut tree = RadixTree::new();

        fn static_handler(_req: &Request, _params: &Params) -> Response {
            Response::ok("static")
        }
        fn param_handler(_req: &Request, _params: &Params) -> Response {
            Response::ok("param")
        }
        fn wildcard_handler(_req: &Request, _params: &Params) -> Response {
            Response::ok("wildcard")
        }

        tree.insert("/users/all", Method::GET, static_handler, "static_handler").unwrap();
        tree.insert("/users/:id", Method::GET, param_handler, "param_handler").unwrap();
        tree.insert("/users/*", Method::GET, wildcard_handler, "wildcard_handler").unwrap();

        let result = tree.lookup("/users/all", &Method::GET);
        assert!(result.is_some());
    }

    #[test]
    fn test_list_routes() {
        let mut tree = RadixTree::new();

        fn handler1(_req: &Request, _params: &Params) -> Response {
            Response::ok("1")
        }
        fn handler2(_req: &Request, _params: &Params) -> Response {
            Response::ok("2")
        }

        tree.insert("/", Method::GET, handler1, "handler1").unwrap();
        tree.insert("/users/:id", Method::GET, handler2, "handler2").unwrap();

        let routes = tree.list_routes();
        assert_eq!(routes.len(), 2);
    }
}

pub trait Compactable {
    fn compact(&mut self) -> std::io::Result<()>;
}

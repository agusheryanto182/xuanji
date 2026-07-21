// pub is for public
// if not define pub, it will be private as default
pub struct User {
    pub first_name: String,
    pub last_name: String,
    pub username: String,
    pub email: String,
    pub age: u8,
}

impl User {
    pub fn say_hello(&self, name: &str) {
        println!("Hello, {}! My name is {}", name, self.first_name);
    }
}

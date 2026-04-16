use crate::third::say_hello_outside as say_hello_outside_third;

pub fn say_hello_outside() {
    println!("Hello from first module outside!");
    say_hello_outside_third();
}

#[test]
fn test_say_hello_outside() {
    say_hello_outside();
}

pub mod second {
    pub mod third {
        pub fn say_hello_outside() {
            // crate::first::say_hello_outside();
            super::super::say_hello_outside();
        }
    }
}

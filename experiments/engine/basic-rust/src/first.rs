use crate::third::say_hello_outside as say_hello_outside_third;

pub fn say_hello_outside() {
    println!("Hello from first module outside!");
    say_hello_outside_third();
}

#[test]
fn test_say_hello_outside() {
    say_hello_outside();
}

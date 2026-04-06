fn main() {
    println!("Hello, world!");
}

#[test]
fn hello_test() {
    println!("Hello test!")
}

#[test]
fn variable() {
    let name = "Suga";
    println!("Name: {}", name);
}

#[test]
fn mutable_variable() {
    let mut count = 0;
    count += 1;
    println!("Count: {}", count);
}

#[test]
fn immutable_variable() {
    let count = 0;

    // comment out to pass the test
    // count += 1; 

    println!("Count: {}", count);
}

#[test]
fn static_typing() {
    let name = "suga";
    println!("Count: {}", name);

    // comment out to pass the test
    // name = 1;
    // println!("Count: {}", name);
}

#[test]
fn shadowing() {
    let name = "suga";
    println!("Name: {}", name);

    let name = name.len();
    println!("Name: {}", name);
}
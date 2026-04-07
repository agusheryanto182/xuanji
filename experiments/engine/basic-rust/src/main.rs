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
fn constant() {
    // const name must be uppercase
    // const must have type annotation
    // const can be declared in any scope
    // const value must be compile-time constant(in other words, it can't receive input from user)
    const NAME: &str = "Suga";
    println!("Name: {}", NAME);
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

#[test]
fn explicit() {
    let name: &str = "suga";
    println!("Name: {}", name);
}

#[test]
fn implicit() {
    let name = "suga";
    println!("Name: {}", name);
}

#[test]
fn number() {
    let x = 1;

    // _ is unused variable
    let _y = 2.0;
    
    println!("x: {}", x);
}

#[test]
fn tuple() {
    let x = (1, 2.0, "suga");
    println!("x: {:?}", x);

    // manual destructuring
    // let a = x.0;
    // let b = x.1;
    // let c = x.2;

    // println!("a: {}, b: {}, c: {}", a, b, c);

    // destructuring
    let (a, _b, c) = x;
    println!("a: {}, c: {}", a, c);
}

#[test]
fn number_conversion() {
    let x : i8 = 10;
    let y : i16 = x as i16;
    
    println!("x: {}, y: {}", x, y);
}

#[test]
fn char_type() {
    let x : char = 'a';
    println!("x: {}", x);
}

// memory management
// Rust is not using garbage collection!
// Rust use stack and heap
// stack is LIFO(Last In First Out)
// heap is FIFO(First In First Out)
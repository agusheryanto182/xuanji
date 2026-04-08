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

#[test]
fn string_slice() {
    let x : &str = "suga";
    let trim = x.trim();        

    println!("x: {}", x);
    println!("trim: {}", trim);
}

#[test]
fn string_type() {
    let x : String = String::from("suga");
    let trim = x.trim();

    println!("x: {}", x);
    println!("trim: {}", trim);
}

#[test]
fn copy_type() {
    let x : i32 = 1;
    let y : i32 = x;

    println!("x: {}", x);
    println!("y: {}", y);
}

#[test]
fn owner_type() {
    let x : String = String::from("suga");
    let y : String = x;

    // comment out, cause the x is moved to y, and x is no longer valid
    // println!("x: {}", x); 
    println!("y: {}", y);
}

#[test]
fn clone_type() {
    let x : String = String::from("suga");
    let y : String = x.clone();

    println!("x: {}", x); 
    println!("y: {}", y);
}

fn _factorial_loop (n: i32) -> i32 {
    if n < 1 {
        return 0;
    }
    
    let mut result = 1;
    
    for i in 1..=n {
        result *= i;
    }

    result
}

#[test]
fn factorial_loop_test() {
    let x = _factorial_loop(5);
    assert_eq!(x, 120);
    println!("x: {}", x);
}

// Rust doesn't use a garbage collector
// Rust uses a stack and a heap

// Stack:
// - LIFO (Last In, First Out)
// - Fast
// - Stores data of a fixed size (known at compile time)

// Stack:
// - Doesn't have a sorted order like FIFO/LIFO
// - More flexible but slower
// - Used for dynamically sized data

// Examples:
// Stack: i32, bool, char, array, tuple
// Heap: String, Vec, HashMap (the original data is on the heap, the pointers are on the stack)

// let s1 = String::from("hello");
// Stack: pointer, length, capacity
// Heap: "hello"

// let s2 = s1;
// Stack s2 pointer to -> "hello"
// Stack s1 pointer is dropped

// but if we clone s1(expensive operation)
// let s2 = s1.clone();
// Stack s2 pointer to -> "hello" (new copy)
// Stack s1 pointer to -> "hello" (original)

// 3 rules about ownership
// each value in Rust must have an owner (owner is variable)
// there can be only one owner at a time
// when owner goes out of scope, the value will be dropped

// ownership in function
fn print_number (number: i32) {
    println!("number: {}", number);
}

fn hi(name: String) {
    println!("Hi, {}", name);
}

#[test]
fn test_ownership() {
    let number = 1;
    print_number(number);
    println!("number: {}", number);

    let name = String::from("suga");
    hi(name);
    // comment out, cause the name is moved to hi, and name is no longer valid
    // println!("name: {}", name);
}

// return value ownership
fn full_name (first_name: String, last_name: String) -> String {
    format!("{} {}", first_name, last_name)
}

#[test]
fn test_return_value_ownership() {
    let first_name = String::from("suga");
    let last_name = String::from("kim");
    let full_name = full_name(first_name, last_name);

    // comment out, cause the first_name and last_name are moved to full_name, and they are no longer valid
    // println!("first_name: {}", first_name);
    // println!("last_name: {}", last_name);

    println!("full_name: {}", full_name);
}

// return the ownership
fn full_name_2 (first_name: String, last_name: String) -> (String, String, String) {
    let full_name = format!("{} {}", first_name, last_name);
    
    (first_name, last_name, full_name)
}

#[test]
fn test_take_ownership() {
    let first_name = String::from("suga");
    let last_name = String::from("kim");
    let (first_name, last_name, full_name) = full_name_2(first_name, last_name);

    println!("first_name: {}", first_name);
    println!("last_name: {}", last_name);
    println!("full_name: {}", full_name);
}

// references and borrowing
fn full_name_3 (first_name: &str, last_name: &str) -> String {
    format!("{} {}", first_name, last_name)
}

#[test]
fn test_references_and_borrowing_1() {
    let first_name = "suga";
    let last_name = "kim";

    let full_name = full_name_3(first_name, last_name);

    println!("first_name: {}", first_name);
    println!("last_name: {}", last_name);
    println!("full_name: {}", full_name);
}

fn full_name_4 (first_name: &String, last_name: &String) -> String {
    format!("{} {}", first_name, last_name)
}

#[test]
fn test_references_and_borrowing_2() {
    let first_name = String::from("suga");
    let last_name = String::from("kim");

    let full_name = full_name_4(&first_name, &last_name);

    println!("first_name: {}", first_name);
    println!("last_name: {}", last_name);
    println!("full_name: {}", full_name);
}

// mutable reference
// 1. At one time, you can have either:
//    - one mutable reference (&mut), OR
//    - any number of immutable references (&)
// 2. To create a mutable reference, the original variable must be declared as `mut`
// 3. While a mutable reference exists, you cannot access the original variable or create other references
// 4. After the mutable reference is no longer used, the original variable can be used again
fn change_value (value: &mut String) {
    value.push_str(" test");
}

#[test]
fn test_mutable_reference() {
    let mut name = String::from("suga");
    change_value(&mut name);
    change_value(&mut name);
    change_value(&mut name);
    
    println!("name: {}", name);
}
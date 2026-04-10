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

// slice

// range
#[test]
fn slice_reference() {
    let array : [i32; 10] = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];
    let slice1 : &[i32] = &array[..];
    let slice2  = &array[0..5];
    let slice3 : &[i32] = &array[5..];
    
    println!("slice1: {:?}", slice1);
    println!("slice2: {:?}", slice2);
    println!("slice3: {:?}", slice3);
}

#[test]
fn string_slice_1() {
    let name = String::from("Raychellz Vermillion");

    let first_name = &name[0..9];
    println!("{}", first_name);

    let last_name = &name[10..20];
    println!("{}", last_name);

    let first_name_1 = first_name;
    println!("{}", first_name_1);

    // note
    // name owns the String data in the heap
    // first_name is a &str slice pointing to "Raychellz" inside name's heap data
    // last_name is a &str slice pointing to "Vermillion" inside name's heap data
    // first_name_1 is a copy of first_name (same pointer + length), pointing to the same slice
}


struct Person {
    first_name: String,
    last_name: String,
    age: u8,
}

// method is a function that is associated with a struct
impl Person {
    fn say_hello(&self, name: &str) {
        println!("Hello, {}! My name is {}", name, self.first_name);
    }
}

fn print_person (person: &Person) {
    println!("first_name: {}", person.first_name);
    println!("last_name: {}", person.last_name);
    println!("age: {}", person.age);
}

// struct
#[test]
fn struct_person() {
    let first_name = String::from("Raychellz");
    let last_name = String::from("Vermillion");

    let person: Person = Person {
        first_name,
        last_name,
        age: 99,
    };

    // println!("first_name: {}", person.first_name);
    // println!("last_name: {}", person.last_name);
    // println!("age: {}", person.age);

    print_person(&person);

    let person2 = Person{..person};
    print_person(&person2);

    // comment out cause the ownership has been moved to person2!
    // solution: clone
    // print_person(&person);

    let person3 = Person{
        first_name: person2.first_name.clone(),
        last_name: person2.last_name.clone(),
        ..person2
    };
    print_person(&person3);
}

// tuple struct
struct GeoPoint(f64, f64);

impl GeoPoint {
    fn new (long: f64, lat: f64) -> GeoPoint {
        GeoPoint(long, lat)
    }
}

#[test]
fn test_associated_function() {
    let geo_point:GeoPoint = GeoPoint::new(1.0, 2.0);
    println!("geo_point: {:?}", geo_point.0);
    println!("geo_point: {:?}", geo_point.1);
}

#[test]
fn tuple_struct() {
    let point = GeoPoint(1.0, 2.0);
    println!("point.0: {}", point.0);
    println!("point.1: {}", point.1);
}

struct Nothing;

#[test]
fn test_nothing(){
    let _nothing1 = Nothing;
    let _nothing2 = Nothing{};
}

#[test]
fn test_method() {
    let person = Person {
        first_name: String::from("Raychellz"),
        last_name: String::from("Vermillion"),
        age: 99
    };

    person.say_hello("Suga");
}

// enum
enum  Level {
    Regular,
    Premium,
    Platinum
}

#[test]
fn test_enum() {
    let _level1 = Level::Platinum;
}

enum Payment {
    CreditCard(String),
    BankTransfer(String, String),
    EWallet (String, String)
}

impl Payment {
    fn pay(&self, amount: u128) {
        println!("Paying amount: {}", amount);
    }
}

#[test]
fn test_payment() {
    let _payment1 = Payment::CreditCard(String::from("21212121"));
    _payment1.pay(1010101);

    let _payment2 = Payment::BankTransfer(String::from("21212121"), String::from("21212121"));
    _payment2.pay(1010101);

    let _payment3 = Payment::EWallet(String::from("21212121"), String::from("21212121"));
    _payment3.pay(10101099991);
}
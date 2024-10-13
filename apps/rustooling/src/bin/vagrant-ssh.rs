#![allow(unused)]

use std::borrow::BorrowMut;
use std::io::prelude::*;
use std::{net::TcpStream, path::Path};

use anyhow::Result;
use clap::Parser;
use rustooling::vagrant::InstanceCollection;

fn main() -> Result<()> {
    let mut args = CliOps::parse();

    let cmd = if let Some(c) = args.cmd.clone() {
        c
    } else {
        "uptime".to_string()
    };

    eprintln!("CliOps: {:?}", args);

    let mut coll = InstanceCollection::from_status();


    /*
    let pk = "/home/jhardy/projects/lab/vagrant-nomad-consul/.vagrant/machines/consul03/virtualbox/private_key";

    let conn = TcpStream::connect("127.0.0.1:2204")?;

    let mut sess = ssh2::Session::new()?;
    sess.set_tcp_stream(conn);
    sess.handshake()?;

    let pk_path = Path::new(pk);

    sess.userauth_pubkey_file("vagrant", None, &pk_path, None)?;

    assert!(sess.authenticated());

    let mut channel = sess.channel_session().unwrap();
    channel.exec(cmd.as_str()).unwrap();
    let mut s = String::new();
    channel.read_to_string(&mut s).unwrap();
    println!("{}", s);
    channel.wait_close();
    println!("{}", channel.exit_status().unwrap());
    */

    Ok(())
}

#[derive(Parser, Debug)]
struct CliOps {
    #[arg(long, short)]
    pub name: Option<Vec<String>>,
    #[arg(long, short)]
    pub cmd: Option<String>,
}

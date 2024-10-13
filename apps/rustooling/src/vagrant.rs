#![allow(unused)]

use anyhow::Result;
use std::{
    borrow::BorrowMut, cell::RefCell, collections::{hash_map, HashMap, HashSet}, io::BufRead, ops::{Deref, DerefMut}, process::Output, rc::Rc
};

#[derive(Debug)]
pub struct VagrantInstance {
    pub name: String,
    meta_data: HashMap<String, String>,
    _ssh_config: Option<SSHConfig>,
}

#[derive(Default, Debug)]
pub struct SSHConfig {
    pub host: Option<String>,
    pub port: Option<String>,
    pub user: Option<String>,
    pub pk: Option<String>,
}

impl VagrantInstance {
    pub fn new(name: String) -> Self {
        Self {
            name,
            meta_data: HashMap::default(),
            _ssh_config: None,
        }
    }

    pub fn add_meta(&mut self, k: String, v: String) {
        self.meta_data.insert(k, v);
    }

    pub fn ssh_config(&mut self) -> Option<&SSHConfig> {
        if self._ssh_config.is_some() {
            return self._ssh_config.as_ref();
        }

        let mut cmd = std::process::Command::new("vagrant")
            .arg("ssh-config")
            .arg(&self.name)
            .output()
            .expect("vagrant ssh-config");

        let lines = cmd.stdout.lines().map(|l| l.unwrap());
        let mut conf = SSHConfig::default();

        for ln in lines.into_iter() {
            let mut ln = ln.trim();

            let mut sp: Vec<_> = ln.split(" ").collect();
            let key = sp[0].to_lowercase();
            let mut value = sp[1..].join(" ");

            match key.to_lowercase().as_str() {
                "user" => {
                    conf.user = Some(value.clone());
                }
                "hostname" => {
                    conf.host = Some(value.clone());
                }
                "port" => {
                    conf.port = Some(value.clone());
                }
                "identityfile" => {
                    conf.pk = Some(value.clone());
                }
                _ => {}
            }
        }

        self._ssh_config = Some(conf);

        self._ssh_config.as_ref()
    }
}

#[derive(Debug)]
pub struct InstanceCollection {
    pub instances: HashMap<String, VagrantInstance>,
}

impl Deref for InstanceCollection {
    type Target = HashMap<String, VagrantInstance>;

    fn deref(&self) -> &Self::Target {
        &self.instances
    }
}

impl DerefMut for InstanceCollection {
    fn deref_mut(&mut self) -> &mut Self::Target {
        &mut self.instances
    }
}

impl InstanceCollection {
    pub fn new() -> Self {
        Self {
            instances: HashMap::new(),
        }
    }

    pub fn from_status() -> Self {
        let mut c = Self::new();

        let mut cmd = std::process::Command::new("vagrant")
            .arg("status")
            .arg("--machine-readable")
            .output()
            .expect("vagrant status result");

        let lines = cmd.stdout.lines().map(|l| l.unwrap());

        // let mut name_set: HashSet<String> = HashSet::new();

        for ln in lines.into_iter() {
            let sp: Vec<_> = ln.split(",").collect();
            // eprintln!("Line: {:?}", sp);

            let iname = sp[1].to_string();
            let mut inst = c.get_instance(iname).expect("instance");

            match sp[2].to_string().as_str() {
                "metadata" => {
                    inst.add_meta(sp[3].to_string(), sp[4].to_string());
                }
                "state-human-short" => {
                    inst.add_meta("state".to_string(), sp[3].to_string());
                }
                _ => {}
            }
        }

        return c;
    }

    pub fn put_instance(&mut self, name: String) -> Result<()> {
        Ok(())
    }

    pub fn get_instance(&mut self, name: String) -> Option<&mut VagrantInstance> {
        Some(
            self.instances
                .entry(name.clone())
                .or_insert(VagrantInstance::new(name.clone())),
        )
    }
}
